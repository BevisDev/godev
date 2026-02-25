package rabbitmq

import "time"

type Option func(*options)

// options defines configuration for RabbitMQ publisher and consumer.
type options struct {
	// prefetchCount limits the number of unacknowledged messages per consumer.
	prefetchCount int

	// autoCommit enables automatic message acknowledgment.
	autoCommit bool

	// publishTimeout sets the timeout for publishing messages.
	publishTimeout time.Duration

	// consumeTimeout sets the timeout for consuming messages.
	consumeTimeout time.Duration

	publisherOn bool
	consumerOn  bool
}

func withDefaults() *options {
	return &options{
		prefetchCount:  10,
		publishTimeout: 5 * time.Second,
		consumeTimeout: 30 * time.Second,
		publisherOn:    true,
		consumerOn:     true,
	}
}

func WithPublisherOnly() Option {
	return func(o *options) {
		o.publisherOn = true
		o.consumerOn = false
	}
}

func WithConsumerOnly() Option {
	return func(o *options) {
		o.consumerOn = true
		o.publisherOn = false
	}
}

func WithPrefetchCount(count int) Option {
	return func(o *options) {
		if count > 0 {
			o.prefetchCount = count
		}
	}
}

func WithAutoCommit() Option {
	return func(o *options) {
		o.autoCommit = true
	}
}

func WithPublishTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout > 0 {
			o.publishTimeout = timeout
		}
	}
}

func WithConsumeTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout > 0 {
			o.consumeTimeout = timeout
		}
	}
}
