package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	seeds         = []string{"redpanda:9092"}
	consumeTopic  = "commands"
	produceTopic  = "replies"
	consumerGroup = "test-server"
)

func main() {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(consumeTopic),
		kgo.DefaultProduceTopic(produceTopic),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)),
	)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	ctx := context.Background()

	// Consuming messages from a topic
	for {
		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			panic(fmt.Sprint(errs))
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			incomingRecord := iter.Next()

			// Extract zilla:correlation-id header
			var correlationID string
			for _, header := range incomingRecord.Headers {
				if header.Key == "zilla:correlation-id" {
					correlationID = string(header.Value)
					break
				}
			}

			// Producing messages
			if correlationID != "" {
				var wg sync.WaitGroup
				wg.Add(1)
				outgoingRecord := &kgo.Record{
					Value: []byte("ok\n"),
					Headers: []kgo.RecordHeader{
						{Key: "zilla:correlation-id", Value: []byte(correlationID)},
					},
				}
				cl.Produce(ctx, outgoingRecord, func(_ *kgo.Record, err error) {
					defer wg.Done()
					if err != nil {
						fmt.Printf("record had a produce error: %v\n", err)
					}
				})
				wg.Wait()
			}
		}
	}
}
