package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"rec-mind/config"
	"rec-mind/internal/database"
	"rec-mind/internal/redis"
	"rec-mind/internal/repository"
	"rec-mind/internal/worker"
	"rec-mind/mq"
)

func main() {
	log.Println("🚀 Starting Query RAG Worker...")

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

	// Test database connection
	ctx := context.Background()
	if err := db.Pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection established")

	// Initialize Redis
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redis.CloseRedis()
	log.Println("✅ Redis connection established")

	// Initialize RabbitMQ
	mq.InitRabbitMQ()
	log.Println("✅ RabbitMQ connection established")

	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db.Pool)

	// Create worker
	queryWorker, err := worker.NewQueryRAGWorker(articleRepo, redis.RedisClient)
	if err != nil {
		log.Fatalf("Failed to create query RAG worker: %v", err)
	}
	log.Println("✅ Query RAG Worker initialized")

	// Start consuming query search jobs
	go func() {
		queue, err := mq.MQChannel.QueueDeclare(
			"query_search_jobs", // name
			true,                // durable
			false,               // delete when unused
			false,               // exclusive
			false,               // no-wait
			nil,                 // arguments
		)
		if err != nil {
			log.Fatalf("Failed to declare query_search_jobs queue: %v", err)
		}

		msgs, err := mq.MQChannel.Consume(
			queue.Name, // queue
			"",         // consumer
			false,      // auto-ack
			false,      // exclusive
			false,      // no-local
			false,      // no-wait
			nil,        // args
		)
		if err != nil {
			log.Fatalf("Failed to register query search jobs consumer: %v", err)
		}

		log.Println("📡 Started consuming query search jobs")

		for d := range msgs {
			var job database.QuerySearchJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				log.Printf("❌ Failed to unmarshal query search job: %v", err)
				d.Nack(false, false)
				continue
			}

			log.Printf("📥 Received query search job %s for query: \"%s\"", job.JobID, job.Query)

			// Process the job
			if err := queryWorker.ProcessQuerySearchJob(job); err != nil {
				log.Printf("❌ Failed to process query search job %s: %v", job.JobID, err)
				d.Nack(false, true) // Requeue on failure
			} else {
				log.Printf("✅ Successfully processed query search job %s", job.JobID)
				d.Ack(false)
			}
		}
	}()

	log.Println("🔍 Query RAG Worker is running... Press Ctrl+C to stop")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Shutting down Query RAG Worker...")
}