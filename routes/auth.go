package routes

import (
	"net/http"

	"admin-site-api/lib"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User is backed by the users table. Password stores a bcrypt hash and is
// never serialized back to clients.
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"column:username"`
	Password string `json:"-" gorm:"column:password"`
}

func (User) TableName() string {
	return "users"
}

// RegisterAuthRoutes wires up POST /login and POST /logout.
func RegisterAuthRoutes(router *gin.Engine, db *gorm.DB) {
	router.POST("/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		var user User
		if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		token, err := lib.GenerateAndStoreToken(db, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Login successful",
			"token":   token,
		})
	})

	router.POST("/logout", lib.AuthMiddleware(db), func(c *gin.Context) {
		token := c.GetString("token")
		username := c.GetString("username")

		result := db.Where("token = ? AND username = ?", token, username).Delete(&lib.Token{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Token removed successfully"})
	})
}
