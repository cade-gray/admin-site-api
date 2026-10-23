package routes

import (
	"net/http"

	"admin-site-api/lib"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User is backed by auth_db's users table (schema:
// https://github.com/cade-gray/jokedle-api/blob/main/auth-db-setup.sql).
// Password stores a bcrypt hash and is never serialized back to clients.
type User struct {
	ID       uint   `json:"id" gorm:"column:idusers;primaryKey"`
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

		// The cookie is what signs the browser in across every app under
		// COOKIE_DOMAIN. The token stays in the body for non-browser clients
		// (scripts, Postman) that send it back as a bearer header.
		if err := lib.IssueSessionCookies(c, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"message":  "Login successful",
			"token":    token,
			"username": user.Username,
		})
	})

	// Who the current session belongs to. The browser apps call this on load:
	// 200 means signed in, 401 means send the visitor to the login page. It is
	// a GET so it needs no CSRF token.
	router.GET("/me", lib.AuthenticateToken(db), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"username": c.GetString("user")})
	})

	router.POST("/logout", lib.AuthenticateToken(db), func(c *gin.Context) {
		token := c.GetString("token")
		username := c.GetString("user")

		result := db.Where("token = ? AND username = ?", token, username).Delete(&lib.Token{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Clear the cookie whether or not a row went away. A token that was
		// already deleted still means this browser should end up signed out.
		lib.ClearSessionCookies(c)

		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "Session already ended"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Token removed successfully"})
	})
}
