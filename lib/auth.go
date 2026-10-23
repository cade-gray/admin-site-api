package lib

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Token represents an issued session token, backed by auth_db's tokens
// table (schema: https://github.com/cade-gray/jokedle-api/blob/main/auth-db-setup.sql).
type Token struct {
	ID        uint      `json:"-" gorm:"column:idtokens;primaryKey"`
	Token     string    `json:"-" gorm:"column:token;uniqueIndex"`
	Username  string    `json:"-" gorm:"column:username"`
	CreatedAt time.Time `json:"-" gorm:"column:createdatetime"`
}

func (Token) TableName() string {
	return "tokens"
}

// GenerateAndStoreToken creates a new random session token for the given
// username and persists it to the tokens table. auth_db's token column is
// VARCHAR(45), so this generates 20 random bytes (40 hex chars).
func GenerateAndStoreToken(db *gorm.DB, username string) (string, error) {
	token, err := randomHex(20)
	if err != nil {
		return "", err
	}

	if err := db.Create(&Token{Token: token, Username: username}).Error; err != nil {
		return "", err
	}
	return token, nil
}

// isValidTokenForUser reports whether tokenString is a currently issued
// token belonging to username.
func isValidTokenForUser(db *gorm.DB, tokenString, username string) bool {
	var record Token
	err := db.Where("token = ? AND username = ?", tokenString, username).First(&record).Error
	return err == nil
}

// AuthenticateToken accepts either of two schemes and puts the caller's
// username and token on the context:
//
//   - Authorization: Bearer <token> plus X-User, the original scheme, still
//     used by scripts, Postman and jokedle-api's clients.
//   - The cg_session cookie, shared across COOKIE_DOMAIN, which is how the
//     browser apps sign in after login.cadegray.dev. The username comes from
//     the token row rather than a header, and state-changing requests must
//     also echo the cg_csrf cookie in X-CSRF-Token.
//
// A request carrying both uses the Authorization header, so an explicit
// credential always wins over whatever the browser happened to attach.
func AuthenticateToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if !strings.HasPrefix(authHeader, "Bearer ") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with Bearer"})
				c.Abort()
				return
			}

			// Extract token string (no UUID validation needed - VARCHAR(45))
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
				c.Abort()
				return
			}

			// Get user from custom header
			user := c.GetHeader("X-User")
			if user == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "X-User header required"})
				c.Abort()
				return
			}

			// Validate token-user combination
			if !isValidTokenForUser(db, tokenString, user) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token for user"})
				c.Abort()
				return
			}

			c.Set("user", user)
			c.Set("token", tokenString)
			c.Next()
			return
		}

		tokenString, err := c.Cookie(SessionCookie)
		if err != nil || tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header or session cookie required"})
			c.Abort()
			return
		}

		if !checkCSRF(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Missing or invalid CSRF token"})
			c.Abort()
			return
		}

		user, ok := userForToken(db, tokenString)
		if !ok {
			// The token is gone from auth_db (logged out elsewhere, or pruned),
			// so clear the stale cookie instead of leaving the browser to retry
			// with it forever.
			ClearSessionCookies(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("token", tokenString)
		c.Next()
	}
}
