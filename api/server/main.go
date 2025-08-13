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
	"rec-mind/config"
	"rec-mind/controllers"
	"rec-mind/handlers"
	"rec-mind/internal/database"
	"rec-mind/internal/mlclient"
	"rec-mind/internal/redis"
	"rec-mind/internal/repository"
	"rec-mind/internal/services"
	"rec-mind/mq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "rec-mind/docs"
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

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db.Pool)
	chunkRepo := repository.NewArticleChunkRepository(db.Pool)

	// Initialize ML client
	mlClient := mlclient.NewMLClient()

	// Initialize article service with ML integration
	articleService := services.NewArticleService(articleRepo, mlClient)

	// Initialize Redis connection
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redis.CloseRedis()

	// Initialize message queue
	mq.InitRabbitMQ()

	// Setup Gin router
	r := gin.Default()

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint (consolidated)
	r.GET("/health", handlers.SystemHealth(db, startTime))

	// API endpoints
	r.POST("/api/upload", uploadArticle(articleService))

	// Article management endpoints
	r.GET("/api/v1/articles", listArticles(articleRepo))
	r.GET("/api/v1/articles/:id", getArticle(articleRepo))
	r.DELETE("/api/v1/articles/:id", deleteArticle(articleRepo))

	// Article chunks endpoints
	r.POST("/api/v1/chunks", createArticleChunk(chunkRepo))
	r.POST("/api/v1/chunks/batch", createArticleChunksBatch(chunkRepo))
	r.GET("/api/v1/chunks", listArticleChunks(chunkRepo))
	r.GET("/api/v1/chunks/:id", getArticleChunk(chunkRepo))
	r.DELETE("/api/v1/chunks/:id", deleteArticleChunk(chunkRepo))
	
	// Article-specific chunks endpoints
	r.GET("/api/v1/articles/:id/chunks", getArticleChunks(chunkRepo))
	r.DELETE("/api/v1/articles/:id/chunks", deleteArticleChunks(chunkRepo))

	// Search endpoints (query-based recommendations)
	searchController := controllers.NewSearchController()
	r.POST("/api/v1/search/recommendations", searchController.SearchByQuery)
	r.POST("/api/v1/search/immediate", searchController.SearchWithImmediateResponse)
	r.GET("/api/v1/search/jobs/:job_id", searchController.GetQuerySearchJobStatus)
	r.GET("/api/v1/search/health", searchController.HealthCheck)

	log.Println("✅ Server ready on :8080")
	log.Println("📚 Swagger UI available at: http://localhost:8080/swagger/index.html")
	r.Run(":8080")
}

// uploadArticle handles article upload with ML embedding generation
// @Summary Upload a new article with ML processing
// @Description Upload a new article to the system and automatically generate embeddings using the Python ML service
// @Tags articles
// @Accept json
// @Produce json
// @Param article body object{title=string,content=string,url=string,category=string} true "Article data"
// @Param processing query string false "Processing mode: 'sync' or 'async' (default: async)"
// @Success 201 {object} object{article=object,message=string,processing_mode=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/upload [post]
func uploadArticle(articleService *services.ArticleService) gin.HandlerFunc {
	return controllers.UploadArticle(articleService)
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

// createArticleChunk handles chunk creation
// @Summary Create a new article chunk
// @Description Create a new chunk for an article with content and metadata
// @Tags chunks
// @Accept json
// @Produce json
// @Param chunk body object{article_id=string,chunk_index=int,content=string,token_count=int,character_count=int} true "Chunk data"
// @Success 201 {object} object{chunk=object}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/chunks [post]
func createArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return controllers.CreateArticleChunk(repo)
}

// createArticleChunksBatch handles batch chunk creation
// @Summary Create multiple article chunks
// @Description Create multiple chunks for an article in a single request
// @Tags chunks
// @Accept json
// @Produce json
// @Param chunks body object{chunks=array} true "Array of chunk data"
// @Success 201 {object} object{chunks=array,count=int}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/chunks/batch [post]
func createArticleChunksBatch(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return controllers.CreateArticleChunksBatch(repo)
}

// listArticleChunks handles chunk listing with pagination
// @Summary List article chunks
// @Description Get a paginated list of article chunks with optional filtering
// @Tags chunks
// @Produce json
// @Param article_id query string false "Filter by article ID"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} object{chunks=array,total=int,limit=int,offset=int}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/chunks [get]
func listArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return controllers.ListArticleChunks(repo)
}

// getArticleChunk handles single chunk retrieval
// @Summary Get article chunk by ID
// @Description Retrieve a single article chunk by its UUID
// @Tags chunks
// @Produce json
// @Param id path string true "Chunk UUID"
// @Success 200 {object} object{chunk=object}
// @Failure 400 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/chunks/{id} [get]
func getArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return controllers.GetArticleChunk(repo)
}


// deleteArticleChunk handles chunk deletion
// @Summary Delete article chunk
// @Description Delete a specific article chunk by its UUID
// @Tags chunks
// @Produce json
// @Param id path string true "Chunk UUID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 404 {object} object{error=string}
// @Router /api/v1/chunks/{id} [delete]
func deleteArticleChunk(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return controllers.DeleteArticleChunk(repo)
}

// getArticleChunks handles retrieving all chunks for an article
// @Summary Get all chunks for an article
// @Description Retrieve all chunks belonging to a specific article
// @Tags chunks
// @Produce json
// @Param id path string true "Article UUID"
// @Success 200 {object} object{chunks=array,count=int}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/articles/{id}/chunks [get]
func getArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return controllers.GetArticleChunks(repo)
}


// deleteArticleChunks handles deleting all chunks for an article
// @Summary Delete all chunks for an article
// @Description Delete all chunks belonging to a specific article
// @Tags chunks
// @Produce json
// @Param id path string true "Article UUID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/articles/{id}/chunks [delete]
func deleteArticleChunks(repo repository.ArticleChunkRepository) gin.HandlerFunc {
	return controllers.DeleteArticleChunks(repo)
}