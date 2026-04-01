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
}

type Consumer struct {
	IsOn bool // enable / disable consumer

	Queue string // queue name

	Handler Handler

	// PrefetchCount sets the AMQP QoS prefetch count (messages) per queue/consumer.
	// If <= 0, it falls back to the MQ default (WithPrefetchCount).
	PrefetchCount int

	// WorkerPool is number of concurrent workers for this consumer.
	// If <= 0, it falls back to 10
	WorkerPool int

	// MaxConsecutiveErrors caps the number of consecutive consume errors before stopping this consumer.
	// If <= 0, it falls back to 10.
	MaxConsecutiveErrors int

	// RetryDelay is the delay between retries after a consume error.
	// If <= 0, it falls back to 5 seconds.
	RetryDelay time.Duration
}
