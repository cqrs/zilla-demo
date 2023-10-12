package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	brokerAddress  = "redpanda:9092"
	commandsTopic  = "commands"
	repliesTopic   = "replies"
)

func main() {
	// Create a new reader for the commands topic
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokerAddress},
		Topic:     commandsTopic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	// Create a new writer for the replies topic
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   repliesTopic,
	})

	for {
		var startTime time.Time

		// Read message from commands topic
		m, err := r.ReadMessage(context.Background())
		if err != nil {
      log.Fatalf("failed to read message: %v", err)
    } else {
			// Capture the start time immediately after reading a message
			startTime = time.Now()
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
