package queue

import "context"

// Message is the unit of data exchanged over the message bus.
type Message struct {
	Key   []byte
	Value []byte
}

// Producer publishes messages to a topic.
type Producer interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

// Consumer subscribes to a topic and dispatches messages to a handler.
type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, msg Message) error) error
	Close() error
}
