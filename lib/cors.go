package lib

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware allows requests from any origin and permits the
// headers/methods this API's routes actually use (Authorization + X-User
// for auth, Content-Type for JSON bodies).
//
// TODO: this is wide open (allow-all) for now - lock it down to specific
// origins before this goes anywhere near production.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, X-User, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
