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

// AuthMiddleware validates the "Authorization: Bearer <token>" header against
// the tokens table. On success it stores "token" and "username" in the gin
// context for downstream handlers.
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
			return
		}

		var record Token
		if err := db.Where("token = ?", parts[1]).First(&record).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("token", record.Token)
		c.Set("username", record.Username)
		c.Next()
	}
}
