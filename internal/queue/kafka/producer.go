package kafka

import (
	"context"
	"fmt"

	"github.com/iPFSoftwares/custos/internal/config"
	"github.com/segmentio/kafka-go"
)

// Producer is a Kafka-backed implementation of queue.Producer.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Producer that writes to the given brokers.
func NewProducer(cfg config.Config) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	return &Producer{writer: w}
}

// Publish sends a single message to the specified topic.
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka producer: publish: %w", err)
	}
	return nil
}

// Close flushes and closes the underlying Kafka writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
