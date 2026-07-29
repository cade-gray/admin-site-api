package lib

import (
	"crypto/rand"
	"encoding/hex"
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
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)

	if err := db.Create(&Token{Token: token, Username: username}).Error; err != nil {
		return "", err
	}
	return token, nil
}


func AuthenticateToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer format
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

		// Store in context
		c.Set("user", user)
		c.Set("token", tokenString)

		c.Next()
	}
}
