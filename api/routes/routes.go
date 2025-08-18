package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"rec-mind/controllers"
	"rec-mind/handlers"
	"rec-mind/internal/database"
	"rec-mind/internal/repository"
	"rec-mind/internal/services"
)

func SetupRoutes(r *gin.Engine, db *database.DB, articleService *services.ArticleService) {
	// CORS configuration for React frontend
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))
	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health
	r.GET("/health", handlers.SystemHealth(db, time.Now()))

	// Repositories
	articleRepo := repository.NewArticleRepository(db.Pool)
	chunkRepo := repository.NewArticleChunkRepository(db.Pool)

	// Controllers
	searchController := controllers.NewSearchController()

	// API routes
	api := r.Group("/api")
	{
		api.POST("/upload", controllers.UploadArticle(articleService))

		v1 := api.Group("/v1")
		{
			// Articles
			articles := v1.Group("/articles")
			{
				articles.GET("", controllers.ListArticles(articleRepo))
				articles.GET("/:id", controllers.GetArticle(articleRepo))
				articles.DELETE("/:id", controllers.DeleteArticle(articleRepo))
				articles.GET("/:id/chunks", controllers.GetArticleChunks(chunkRepo))
				articles.DELETE("/:id/chunks", controllers.DeleteArticleChunks(chunkRepo))
			}

			// Chunks
			chunks := v1.Group("/chunks")
			{
				chunks.POST("", controllers.CreateArticleChunk(chunkRepo))
				chunks.POST("/batch", controllers.CreateArticleChunksBatch(chunkRepo))
				chunks.GET("", controllers.ListArticleChunks(chunkRepo))
				chunks.GET("/:id", controllers.GetArticleChunk(chunkRepo))
				chunks.DELETE("/:id", controllers.DeleteArticleChunk(chunkRepo))
			}

			// Search
			search := v1.Group("/search")
			{
				search.POST("/recommendations", searchController.SearchByQuery)
				search.POST("/immediate", searchController.SearchWithImmediateResponse)
				search.GET("/jobs/:job_id", searchController.GetQuerySearchJobStatus)
			}
		}
	}
}