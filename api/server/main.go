// @title RecMind API
// @version 1.0
// @description A distributed news article recommendation system API with ML integration
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://github.com/MartinHeinz/go-project-blueprint/blob/master/LICENSE

// @host localhost:8080
// @BasePath /
// @schemes http

package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owenHochwald/rec-mind-api/config"
	"github.com/owenHochwald/rec-mind-api/controllers"
	"github.com/owenHochwald/rec-mind-api/handlers"
	"github.com/owenHochwald/rec-mind-api/internal/database"
	"github.com/owenHochwald/rec-mind-api/internal/repository"
	"github.com/owenHochwald/rec-mind-api/mq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/owenHochwald/rec-mind-api/docs"
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

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoints
	r.GET("/health", handlers.SystemHealth(db, startTime))
	r.GET("/health/detail", handlers.DetailedHealth(db, startTime))
	r.GET("/health/python", handlers.CheckPythonHealth())

	// API endpoints
	r.POST("/api/upload", uploadArticle(articleRepo))
	r.POST("/api/interact", handleInteraction())
	r.GET("/api/recommend", getRecommendations())

	// Article management endpoints
	r.GET("/api/v1/articles", listArticles(articleRepo))
	r.GET("/api/v1/articles/:id", getArticle(articleRepo))
	r.PUT("/api/v1/articles/:id", updateArticle(articleRepo))
	r.DELETE("/api/v1/articles/:id", deleteArticle(articleRepo))

	log.Println("✅ Server ready on :8080")
	log.Println("📚 Swagger UI available at: http://localhost:8080/swagger/index.html")
	r.Run(":8080")
}

// uploadArticle handles article upload with message queue publishing
// @Summary Upload a new article
// @Description Upload a new article to the system and publish it to the message queue for ML processing
// @Tags articles
// @Accept json
// @Produce json
// @Param article body object{title=string,content=string,url=string,category=string} true "Article data"
// @Success 201 {object} object{message=string,article_id=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/upload [post]
func uploadArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return controllers.UploadArticleV2(repo)
}

// handleInteraction handles user interactions (placeholder)
// @Summary Handle user interaction
// @Description Process user interaction data (placeholder endpoint)
// @Tags interactions
// @Accept json
// @Produce json
// @Success 200 {object} object{message=string}
// @Router /api/interact [post]
func handleInteraction() gin.HandlerFunc {
	return controllers.HandleInteraction
}

// getRecommendations handles recommendation requests (placeholder)
// @Summary Get article recommendations
// @Description Get personalized article recommendations (placeholder endpoint)
// @Tags recommendations
// @Produce json
// @Success 200 {object} object{message=string}
// @Router /api/recommend [get]
func getRecommendations() gin.HandlerFunc {
	return controllers.GetRecommendations
}

// listArticles handles article listing with pagination
// @Summary List articles
// @Description Get a paginated list of articles with optional filtering
// @Tags articles
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param category query string false "Filter by category"
// @Success 200 {object} object{articles=array,pagination=object}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/articles [get]
func listArticles(repo repository.ArticleRepository) gin.HandlerFunc {
	return controllers.ListArticles(repo)
}

// getArticle handles single article retrieval
// @Summary Get article by ID
// @Description Retrieve a single article by its UUID
// @Tags articles
// @Produce json
// @Param id path string true "Article UUID"
// @Success 200 {object} object{article=object}
// @Failure 400 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/articles/{id} [get]
func getArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return controllers.GetArticle(repo)
}

// updateArticle handles article updates
// @Summary Update article
// @Description Update an existing article by its UUID
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article UUID"
// @Param article body object{title=string,content=string,url=string,category=string} true "Updated article data"
// @Success 200 {object} object{message=string,article=object}
// @Failure 400 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/articles/{id} [put]
func updateArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return controllers.UpdateArticle(repo)
}

// deleteArticle handles article deletion
// @Summary Delete article
// @Description Delete an article by its UUID
// @Tags articles
// @Produce json
// @Param id path string true "Article UUID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/articles/{id} [delete]
func deleteArticle(repo repository.ArticleRepository) gin.HandlerFunc {
	return controllers.DeleteArticle(repo)
}