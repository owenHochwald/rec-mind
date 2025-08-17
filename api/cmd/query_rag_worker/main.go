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
	"rec-mind/models"
	"rec-mind/mq"
)

func main() {
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("X Failed to load database config: %v", err)
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("X Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("- Database connection established")

	if err := redis.InitRedis(); err != nil {
		log.Fatalf("X Failed to initialize Redis: %v", err)
	}
	defer redis.CloseRedis()
	log.Println("- Redis connection established")

	mq.InitRabbitMQ()
	log.Println("- RabbitMQ connection established")

	articleRepo := repository.NewArticleRepository(db.Pool)

	queryWorker, err := worker.NewQueryRAGWorker(articleRepo, redis.RedisClient)
	if err != nil {
		log.Fatalf("Failed to create query RAG worker: %v", err)
	}
	log.Println("✅ Query RAG Worker initialized")

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
			var job models.QuerySearchJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				log.Printf("X Failed to unmarshal query search job: %v", err)
				d.Nack(false, false)
				continue
			}

			log.Printf("📥 Received query search job %s for query: \"%s\"", job.JobID, job.Query)

			if err := queryWorker.ProcessQuerySearchJob(job); err != nil {
				log.Printf("X Failed to process query search job %s: %v", job.JobID, err)
				d.Nack(false, true) // Requeue on failure
			} else {
				log.Printf("✅ Successfully processed query search job %s", job.JobID)
				d.Ack(false)
			}
		}
	}()

	log.Println("Query RAG Worker is running... Press Ctrl+C to stop")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}