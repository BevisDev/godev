package rabbitmq

import (
	"context"
	"time"
)

// Handler defines the interface for message consumers.
type Handler interface {
	// Handle processes a single message. Returns nil to ack, error to requeue.
	// When autoCommit is false, handler must call msg.Commit() on success.
	Handle(ctx context.Context, msg *MsgHandler) error

	// QueueName returns the AMQP queue this handler consumes from.
	QueueName() string
}

type Consumer struct {
	Handler Handler

	IsOn bool // enable / disable consumer

	// PrefetchCount sets the AMQP QoS prefetch count (messages) per queue/consumer.
	// If <= 0, CM uses its default (prefetch 1).
	PrefetchCount int

	// WorkerPool is number of concurrent workers for this consumer.
	// If <= 0, CM uses its default (10).
	WorkerPool int

	// MaxConsecutiveErrors caps the number of consecutive consume errors before stopping this consumer.
	// If <= 0, it falls back to 10.
	MaxConsecutiveErrors int

	// RetryDelay is the delay between retries after a consume error.
	// If <= 0, it falls back to 5 seconds.
	RetryDelay time.Duration
}
