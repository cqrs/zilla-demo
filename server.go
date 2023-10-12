package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"os"

	"github.com/segmentio/kafka-go"
)

const (
	brokerAddress  = "redpanda:9092"
	commandsTopic  = "commands"
	repliesTopic   = "replies"
)

func main() {
	logger := log.New(os.Stdout, "[kafka] ", 0)
	// Create a new reader for the commands topic
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerAddress},
		Topic:     commandsTopic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6, // 10MB
		MaxWait:   60000 * time.Millisecond,
		Logger:    logger,
	})

	// Create a new writer for the replies topic
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{brokerAddress},
		Topic:        repliesTopic,
		BatchSize:    100,
    BatchTimeout: 10 * time.Millisecond,
		Logger:       logger,
	})

	for {
		var startTime time.Time

		// Read message from commands topic
		m, err := r.ReadMessage(context.Background())
		if err != nil {
      log.Fatalf("failed to read message: %v", err)
    }

		// Extract zilla:correlation-id header
		var correlationID string
		for _, header := range m.Headers {
			if header.Key == "zilla:correlation-id" {
				correlationID = string(header.Value)
				break
			}
		}

		if correlationID != "" {
			// Create and write response message to replies topic
			responseHeaders := []kafka.Header{
				{Key: "zilla:correlation-id", Value: []byte(correlationID)},
			}
			startTime = time.Now()
			responseMessage := kafka.Message{
				Value:   []byte("ok"),
				Headers: responseHeaders,
			}
			if err := w.WriteMessages(context.Background(), responseMessage); err != nil {
				log.Fatalf("failed to write message: %v", err)
			}

			// Capture the end time and calculate the duration
			endTime := time.Now()
			duration := endTime.Sub(startTime)

			// Log the duration
			fmt.Printf("Sent response for correlation ID %s, duration: %v\n", correlationID, duration)
		}
	}
}
