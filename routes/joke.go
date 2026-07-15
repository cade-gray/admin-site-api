package routes

import (
	"encoding/json"
	"net/http"
	"os"

	"admin-site-api/lib"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterJokeRoutes wires up POST /joke, which pulls a random joke from the
// RapidAPI dad-jokes API. Requires a valid bearer token (see lib.AuthMiddleware)
// and the RAPIDAPI_KEY environment variable.
func RegisterJokeRoutes(router *gin.Engine, db *gorm.DB) {
	router.POST("/joke", lib.AuthMiddleware(db), func(c *gin.Context) {
		apiKey := os.Getenv("RAPIDAPI_KEY")
		if apiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "RAPIDAPI_KEY is not configured"})
			return
		}

		req, err := http.NewRequest(http.MethodGet, "https://dad-jokes.p.rapidapi.com/random/joke", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
			return
		}
		req.Header.Set("X-RapidAPI-Key", apiKey)
		req.Header.Set("X-RapidAPI-Host", "dad-jokes.p.rapidapi.com")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch joke"})
			return
		}
		defer resp.Body.Close()

		var result struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Body    []struct {
				Setup     string `json:"setup"`
				Punchline string `json:"punchline"`
			} `json:"body"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode joke"})
			return
		}

		if !result.Success || len(result.Body) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Message})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"setup":     result.Body[0].Setup,
			"punchline": result.Body[0].Punchline,
		})
	})
}
