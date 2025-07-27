package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owenHochwald/rec-mind-api/config"
	"github.com/owenHochwald/rec-mind-api/controllers"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/health"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
	"github.com/owenHochwald/rec-mind-api/mq"
)

func main() {
	startTime := time.Now()
	log.Println("🚀 Starting rec-mind API server...")

	// Load database configuration
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repository
	articleRepo := repository.NewArticleRepository(db.Pool)

	// Initialize message queue
	mq.InitRabbitMQ()

	// Setup Gin router
	r := gin.Default()

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		systemHealth := health.CheckSystemHealth(db, startTime)
		if systemHealth.Status == "healthy" {
			c.JSON(200, systemHealth)
		} else {
			c.JSON(503, systemHealth)
		}
	})

	r.GET("/health/detail", func(c *gin.Context) {
		dbHealth := health.CheckDatabase(db)
		c.JSON(200, gin.H{
			"database": dbHealth,
			"uptime":   time.Since(startTime).String(),
			"version":  "1.0.0",
		})
	})

	// API endpoints
	r.POST("/api/upload", controllers.UploadArticleV2(articleRepo))
	r.POST("/api/interact", controllers.HandleInteraction)
	r.GET("/api/recommend", controllers.GetRecommendations)

	// Article management endpoints
	r.GET("/api/v1/articles", controllers.ListArticles(articleRepo))
	r.GET("/api/v1/articles/:id", controllers.GetArticle(articleRepo))
	r.PUT("/api/v1/articles/:id", controllers.UpdateArticle(articleRepo))
	r.DELETE("/api/v1/articles/:id", controllers.DeleteArticle(articleRepo))

	log.Println("✅ Server ready on :8080")
	r.Run(":8080")
}
