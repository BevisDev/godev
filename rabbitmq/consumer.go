package rabbitmq

import (
	"context"
	"time"
)

const (
	defaultMaxConsecutiveErrors = 10
	defaultPrefetchCount        = 10
	defaultWorkerPool           = 10
	defaultRetryDelay           = 5 * time.Second
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
	// IsOn is the flag to enable / disable consume
	IsOn bool

	// Handler defines the interface for message consumers.
	Handler Handler

	// Options is option for consumer
	Options ConsumerOptions
}

type ConsumerOptions struct {
	// PrefetchCount sets the AMQP QoS prefetch count (messages) per queue/consumer.
	// If <= 0, it falls back to 1
	PrefetchCount int

	// WorkerPool is number of concurrent workers for this consumer.
	// If <= 0, it falls back to 10.
	WorkerPool int

	// MaxConsecutiveErrors caps the number of consecutive consume errors before stopping this consumer.
	// If <= 0, it falls back to 10.
	MaxConsecutiveErrors int

	// RetryDelay is the delay between retries after a consume error.
	// If <= 0, it falls back to 5 seconds.
	RetryDelay time.Duration

	// BatchSize flushes the batch when this many messages are collected.
	BatchSize int

	// FlushInterval flushes the batch after this duration even if BatchSize
	FlushInterval time.Duration
}

func (c *ConsumerOptions) withDefaults() {
	if c.PrefetchCount <= 0 {
		c.PrefetchCount = defaultPrefetchCount
	}
	if c.WorkerPool <= 0 {
		c.WorkerPool = defaultWorkerPool
	}
	if c.MaxConsecutiveErrors <= 0 {
		c.MaxConsecutiveErrors = defaultMaxConsecutiveErrors
	}
	if c.RetryDelay <= 0 {
		c.RetryDelay = defaultRetryDelay
	}
}
