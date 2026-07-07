package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/iPFSoftwares/custos/internal/config"
	"github.com/iPFSoftwares/custos/internal/queue"
)

// Compile-time interface check.
var _ queue.Consumer = (*Consumer)(nil)

// Consumer is a Kafka-backed implementation of queue.Consumer.
type Consumer struct {
	brokers []string
}

// NewConsumer creates a Consumer connected to the configured brokers.
func NewConsumer(cfg config.Config) *Consumer {
	return &Consumer{brokers: cfg.Kafka.Brokers}
}

// Subscribe reads messages from topic in a loop, calling handler for each.
// Messages are committed only after handler returns nil. Blocks until ctx is cancelled.
func (c *Consumer) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, msg queue.Message) error) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  "custos-server",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer r.Close()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // context cancelled — normal shutdown
			}
			return fmt.Errorf("kafka consumer: fetch: %w", err)
		}

		if err := handler(ctx, queue.Message{Key: m.Key, Value: m.Value}); err != nil {
			// Log but do not commit; message will be redelivered.
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("kafka consumer: commit: %w", err)
		}
	}
}

// Close releases any consumer-level resources. Individual readers are closed
// inside Subscribe; this is a no-op placeholder to satisfy the interface.
func (c *Consumer) Close() error {
	return nil
}
