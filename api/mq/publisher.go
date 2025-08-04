package mq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"fmt"

	"github.com/joho/godotenv"
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

	MQConn = conn
	MQChannel = ch
	log.Println("🐰 RabbitMQ connection initialized")
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
