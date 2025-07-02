package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/owenHochwald/rec-mind-api/models"
)

// adds an article from JSON recieved in the request body to the database
func UploadArticle(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newArticle models.Article

		// use BindJSON to bind json to newArticle
		if err := c.BindJSON(&newArticle); err != nil {
			return
		}

		query := `
			INSERT INTO articles (title, content, tags)
			VALUES ($1, $2, $3)
			`

		_, err := db.Exec(query, newArticle.Title, newArticle.Content, pq.Array(newArticle.Tags))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert article"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Article uploaded successfully"})
	}
}
func HandleInteraction(c *gin.Context)  {}
func GetRecommendations(c *gin.Context) {}
