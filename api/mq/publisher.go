package mq

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/joho/godotenv"
	"rec-mind/models"
)

var MQConn *amqp.Connection
var MQChannel *amqp.Channel

func InitRabbitMQ() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPassword := os.Getenv("RABBITMQ_PASSWORD")

	amqpURL := fmt.Sprintf("amqp://%s:%s@localhost:5672/", rabbitUser, rabbitPassword)

	conn, err := amqp.Dial(amqpURL)
	
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	// Declare article_events queue
	_, err = ch.QueueDeclare(
		"article_events", // queue name
		true,             // durable
		false,            // auto-delete
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare article_events queue: %v", err)
	}

	// Declare article_processing queue
	_, err = ch.QueueDeclare(
		"article_processing", // queue name
		true,                 // durable
		false,                // auto-delete
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare article_processing queue: %v", err)
	}

	// Declare recommendation_jobs queue
	_, err = ch.QueueDeclare(
		"recommendation_jobs", // queue name
		true,                  // durable
		false,                 // auto-delete
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare recommendation_jobs queue: %v", err)
	}

	// Declare chunk_search queue
	_, err = ch.QueueDeclare(
		"chunk_search", // queue name
		true,           // durable
		false,          // auto-delete
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare chunk_search queue: %v", err)
	}

	// Declare search_results queue
	_, err = ch.QueueDeclare(
		"search_results", // queue name
		true,             // durable
		false,            // auto-delete
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare search_results queue: %v", err)
	}

	// Declare query_search_jobs queue
	_, err = ch.QueueDeclare(
		"query_search_jobs", // queue name
		true,                // durable
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare query_search_jobs queue: %v", err)
	}

	// Declare query_search queue
	_, err = ch.QueueDeclare(
		"query_search", // queue name
		true,           // durable
		false,          // auto-delete
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare query_search queue: %v", err)
	}

	MQConn = conn
	MQChannel = ch
}

func PublishEvent(body string) error {
	err := MQChannel.Publish(
		"",               // exchange
		"article_events", // routing key (queue name)
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	return err
}

// PublishQuerySearchJob publishes a query search job to the jobs queue
func PublishQuerySearchJob(job models.QuerySearchJob) error {
	messageBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal query search job: %w", err)
	}

	err = MQChannel.Publish(
		"",                  // exchange
		"query_search_jobs", // routing key (queue name)
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBytes,
			DeliveryMode: 2, // persistent
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish query search job: %w", err)
	}

	log.Printf("📤 Published query search job %s for query: %s", job.JobID, job.Query)
	return nil
}

// PublishQuerySearch publishes a query search message to the search queue
func PublishQuerySearch(message models.QuerySearchMessage) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal query search message: %w", err)
	}

	err = MQChannel.Publish(
		"",             // exchange
		"query_search", // routing key (queue name)
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBytes,
			DeliveryMode: 2, // persistent
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish query search message: %w", err)
	}

	log.Printf("📤 Published query search %s for query: %s", message.SearchID, message.Query)
	return nil
}

// PublishArticleProcessing publishes an article to the processing queue for chunking and embedding
func PublishArticleProcessing(articleID, title, content, category string, createdAt string) error {
	message := map[string]interface{}{
		"article_id": articleID,
		"title":      title,
		"content":    content,
		"category":   category,
		"created_at": createdAt,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal article processing message: %w", err)
	}

	err = MQChannel.Publish(
		"",                   // exchange
		"article_processing", // routing key (queue name)
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBytes,
			DeliveryMode: 2, // Make message persistent (2 = persistent, 1 = transient)
		},
	)
	
	if err != nil {
		return fmt.Errorf("failed to publish article processing message: %w", err)
	}
	
	log.Printf("📨 Published article %s to processing queue", articleID)
	return nil
}

// PublishRecommendationJob publishes a recommendation job to the jobs queue
func PublishRecommendationJob(job models.RecommendationJob) error {
	messageBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendation job: %w", err)
	}

	err = MQChannel.Publish(
		"",                    // exchange
		"recommendation_jobs", // routing key (queue name)
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBytes,
			DeliveryMode: 2, // persistent
		},
	)
	
	if err != nil {
		return fmt.Errorf("failed to publish recommendation job: %w", err)
	}
	
	log.Printf("📨 Published recommendation job %s", job.JobID)
	return nil
}

// PublishChunkSearch publishes a chunk search message to the search queue
func PublishChunkSearch(message models.ChunkSearchMessage) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk search message: %w", err)
	}

	err = MQChannel.Publish(
		"",            // exchange
		"chunk_search", // routing key (queue name)
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBytes,
			DeliveryMode: 2, // persistent
		},
	)
	
	if err != nil {
		return fmt.Errorf("failed to publish chunk search: %w", err)
	}
	
	log.Printf("📨 Published chunk search %s for job %s", message.SearchID, message.JobID)
	return nil
}
