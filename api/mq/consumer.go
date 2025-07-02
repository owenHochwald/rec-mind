package mq

import "log"

func StartConsumer() {
	msgs, err := MQChannel.Consume(
		"article_events", // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			log.Printf("📨 Received message: %s", string(msg.Body))
			// Add future: generate recs, update user cache, etc
		}
	}()
}
