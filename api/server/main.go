package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"rec-mind/config"
	"rec-mind/internal/database"
	"rec-mind/internal/mlclient"
	"rec-mind/internal/redis"
	"rec-mind/internal/repository"
	"rec-mind/internal/services"
	"rec-mind/mq"
	"rec-mind/routes"
	_ "rec-mind/docs"
)

func main() {
	startTime := time.Now()

	db := initializeDatabase()
	defer db.Close()

	initializeRedis()
	defer redis.CloseRedis()

	mq.InitRabbitMQ()

	r := gin.Default()
	articleService := initializeServices(db)
	
	routes.SetupRoutes(r, db, articleService)

	log.Printf("Server ready on :8080 (startup: %v)", time.Since(startTime))
	log.Println("API Documentation: http://localhost:8080/swagger/index.html")
	
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initializeDatabase() *database.DB {
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected")
	return db
}

func initializeRedis() {
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("Redis connected")
}

func initializeServices(db *database.DB) *services.ArticleService {
	articleRepo := repository.NewArticleRepository(db.Pool)
	mlClient := mlclient.NewMLClient()
	return services.NewArticleService(articleRepo, mlClient)
}