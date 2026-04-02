package rabbitmq

type Option func(*options)

const (
	defaultReconnectMaxRetries = 10
)

// options defines configuration for RabbitMQ producer and consumer.
type options struct {
	// autoCommit enables automatic message acknowledgment.
	autoCommit bool

	// reconnectMaxRetries sets max attempts for reconnect.
	reconnectMaxRetries int

	producerOn bool
	consumerOn bool
}

func withDefaults() *options {
	return &options{
		reconnectMaxRetries: defaultReconnectMaxRetries,
		producerOn:          true,
		consumerOn:          true,
	}
}

func WithProducerOnly() Option {
	return func(o *options) {
		o.producerOn = true
		o.consumerOn = false
	}
}

func WithConsumerOnly() Option {
	return func(o *options) {
		o.consumerOn = true
		o.producerOn = false
	}
}

func WithAutoCommit() Option {
	return func(o *options) {
		o.autoCommit = true
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
