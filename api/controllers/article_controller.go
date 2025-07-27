package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
	"github.com/owenHochwald/rec-mind-api/models"
	"github.com/owenHochwald/rec-mind-api/mq"
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

		jsonData, _ := json.Marshal(newArticle)
		err = mq.PublishEvent(string(jsonData))
		if err != nil {
			log.Println("Failed to publish message:", err)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article uploaded successfully"})
	}
}
// UploadArticleV2 creates a new article using the repository pattern
func UploadArticleV2(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req database.CreateArticleRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		article, err := repo.Create(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
			return
		}

		// Publish to message queue
		jsonData, _ := json.Marshal(article)
		if err := mq.PublishEvent(string(jsonData)); err != nil {
			log.Printf("Failed to publish article event: %v", err)
		}

		c.JSON(http.StatusCreated, article.ToResponse())
	}
}

// ListArticles retrieves articles with filtering and pagination
func ListArticles(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter database.ArticleFilter

		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter.SetDefaults()
		articles, err := repo.List(c.Request.Context(), &filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve articles"})
			return
		}

		count, err := repo.Count(c.Request.Context(), &filter)
		if err != nil {
			log.Printf("Failed to get article count: %v", err)
			count = 0
		}

		response := gin.H{
			"articles": articles,
			"total":    count,
			"limit":    filter.Limit,
			"offset":   filter.Offset,
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetArticle retrieves a single article by ID
func GetArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		article, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		c.JSON(http.StatusOK, article.ToResponse())
	}
}

// UpdateArticle updates an existing article
func UpdateArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		var req database.UpdateArticleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		article, err := repo.Update(c.Request.Context(), id, &req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		c.JSON(http.StatusOK, article.ToResponse())
	}
}

// DeleteArticle removes an article
func DeleteArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
			return
		}

		if err := repo.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
	}
}

func HandleInteraction(c *gin.Context)  {}
func GetRecommendations(c *gin.Context) {}
