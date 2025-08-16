package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"rec-mind/internal/database"
	"rec-mind/internal/redis"
	"rec-mind/internal/repository"
	"rec-mind/mq"
)

type JobConsumer struct {
	queryWorker *QueryRAGWorker
	channel     *amqp.Channel
	isRunning   bool
	wg          sync.WaitGroup
}

func NewJobConsumer(chunkRepo repository.ArticleChunkRepository, articleRepo repository.ArticleRepository) (*JobConsumer, error) {
	if mq.MQChannel == nil {
		return nil, fmt.Errorf("RabbitMQ channel not initialized")
	}

	if redis.RedisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	queryWorker, err := NewQueryRAGWorker(articleRepo, redis.RedisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create query RAG worker: %w", err)
	}

	return &JobConsumer{
		queryWorker: queryWorker,
		channel:     mq.MQChannel,
		isRunning:   false,
	}, nil
}

func (jc *JobConsumer) Start() error {
	if jc.isRunning {
		return fmt.Errorf("job consumer is already running")
	}

	// Declare the query_search_jobs queue
	queue, err := jc.channel.QueueDeclare(
		"query_search_jobs", // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare query_search_jobs queue: %w", err)
	}

	// Set QoS to process one job at a time per worker
	err = jc.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := jc.channel.Consume(
		queue.Name,            // queue
		"query-rag-worker",    // consumer
		false,                 // auto-ack
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register query search jobs consumer: %w", err)
	}

	jc.isRunning = true
	jc.wg.Add(1)

	go func() {
		defer jc.wg.Done()
		log.Println("🚀 Started query search jobs consumer")

		for d := range msgs {
			if !jc.isRunning {
				break
			}

			var job database.QuerySearchJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				log.Printf("❌ Failed to unmarshal query search job: %v", err)
				d.Nack(false, false)
				continue
			}

			log.Printf("📋 Processing query search job %s for query: %s", job.JobID, job.Query)

			// Process the job using query RAG worker
			if err := jc.queryWorker.ProcessQuerySearchJob(job); err != nil {
				log.Printf("❌ Failed to process query search job %s: %v", job.JobID, err)
				d.Nack(false, true) // Requeue for retry
				continue
			}

			// Acknowledge successful processing
			d.Ack(false)
			log.Printf("✅ Successfully processed query search job %s", job.JobID)
		}

		log.Println("🛑 Query search jobs consumer stopped")
	}()

	return nil
}

func (jc *JobConsumer) Stop() {
	if !jc.isRunning {
		return
	}

	log.Println("🛑 Stopping query search jobs consumer...")
	jc.isRunning = false
	jc.wg.Wait()
	log.Println("✅ Query search jobs consumer stopped")
}

func (jc *JobConsumer) IsRunning() bool {
	return jc.isRunning
}