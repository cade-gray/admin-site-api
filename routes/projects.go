package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Project is backed by the projects table.
type Project struct {
	ProjectID   int    `json:"projectId" gorm:"column:project_id;primaryKey"`
	Name        string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
}

func (Project) TableName() string {
	return "projects"
}

// RegisterProjectRoutes wires up GET /projects.
func RegisterProjectRoutes(router *gin.Engine, db *gorm.DB) {
	router.GET("/projects", func(c *gin.Context) {
		var projects []Project
		if err := db.Order("project_id desc").Find(&projects).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
			return
		}
		c.JSON(http.StatusOK, projects)
	})
}
