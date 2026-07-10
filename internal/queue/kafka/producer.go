package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/Godie360/custos/internal/config"
	"github.com/Godie360/custos/internal/queue"
)

// Compile-time interface check.
var _ queue.Producer = (*Producer)(nil)

// ProducerOption configures a Kafka Producer.
type ProducerOption func(*kafka.Writer)

// WithAsync switches the writer to async mode (higher throughput, no per-write ack).
func WithAsync() ProducerOption {
	return func(w *kafka.Writer) { w.Async = true }
}

// WithWriteTimeout sets the per-write deadline. Default: 10s.
func WithWriteTimeout(d time.Duration) ProducerOption {
	return func(w *kafka.Writer) { w.WriteTimeout = d }
}

// WithReadTimeout sets the per-read deadline. Default: 10s.
func WithReadTimeout(d time.Duration) ProducerOption {
	return func(w *kafka.Writer) { w.ReadTimeout = d }
}

// Producer is a Kafka-backed implementation of queue.Producer.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Producer that writes to the brokers from cfg.
func NewProducer(cfg config.Config, opts ...ProducerOption) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
	}
	for _, opt := range opts {
		opt(w)
	}
	return &Producer{writer: w}
}

// Publish sends a single message to the specified topic.
// ctx controls the write deadline — callers SHOULD set a timeout.
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka producer: publish to %q: %w", topic, err)
	}
	return nil
}

// Close flushes and closes the underlying Kafka writer.
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("kafka producer: close: %w", err)
	}
	return nil
}
