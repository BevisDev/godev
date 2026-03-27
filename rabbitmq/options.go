package rabbitmq

import "time"

type Option func(*options)

// options defines configuration for RabbitMQ publisher and consumer.
type options struct {
	// autoCommit enables automatic message acknowledgment.
	autoCommit bool

	// publishTimeout sets the timeout for publishing messages.
	publishTimeout time.Duration

	// consumeTimeout sets the timeout for consuming messages.
	consumeTimeout time.Duration

	// reconnectMaxRetries sets max attempts for reconnect.
	reconnectMaxRetries int

	publisherOn bool
	consumerOn  bool
}

func withDefaults() *options {
	return &options{
		publishTimeout:      5 * time.Second,
		consumeTimeout:      30 * time.Second,
		reconnectMaxRetries: 10,
		publisherOn:         true,
		consumerOn:          true,
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

// WithReconnectMaxRetries sets max retry attempts when reconnecting.
func WithReconnectMaxRetries(maxRetries int) Option {
	return func(o *options) {
		if maxRetries > 0 {
			o.reconnectMaxRetries = maxRetries
		}
	}
}
