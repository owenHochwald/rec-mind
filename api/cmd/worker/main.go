package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rec-mind/config"
	"rec-mind/internal/database"
	"rec-mind/internal/redis"
	"rec-mind/internal/repository"
	"rec-mind/internal/worker"
	"rec-mind/mq"
)

func main() {
	log.Println("🚀 Starting RAG Worker Service")

	// Initialize database connection
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load database config: %v", err)
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Redis connection
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("❌ Failed to initialize Redis: %v", err)
	}
	defer func() {
		if err := redis.CloseRedis(); err != nil {
			log.Printf("⚠️ Error closing Redis: %v", err)
		}
	}()

	// Initialize RabbitMQ connection
	mq.InitRabbitMQ()
	defer func() {
		if mq.MQChannel != nil {
			mq.MQChannel.Close()
		}
		if mq.MQConn != nil {
			mq.MQConn.Close()
		}
	}()

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db.Pool)
	chunkRepo := repository.NewArticleChunkRepository(db.Pool)

	// Create and start job consumer
	jobConsumer, err := worker.NewJobConsumer(chunkRepo, articleRepo)
	if err != nil {
		log.Fatalf("❌ Failed to create job consumer: %v", err)
	}

	if err := jobConsumer.Start(); err != nil {
		log.Fatalf("❌ Failed to start job consumer: %v", err)
	}

	// Health check goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			
			// Check Redis health
			if err := redis.HealthCheck(ctx); err != nil {
				log.Printf("⚠️ Redis health check failed: %v", err)
			}

			// Check database health
			if err := db.Pool.Ping(ctx); err != nil {
				log.Printf("⚠️ Database health check failed: %v", err)
			}

			cancel()
		}
	}()

	log.Println("✅ RAG Worker Service started successfully")
	log.Println("📋 Listening for recommendation jobs...")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down RAG Worker Service...")

	// Graceful shutdown
	jobConsumer.Stop()

	log.Println("✅ RAG Worker Service stopped gracefully")
}