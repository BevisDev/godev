package kafkax

import "time"

// Option configures Kafka client construction (mirrors rabbitmq.Option pattern).
type Option func(*options)

type options struct {
	// autoCommit sets Consumer.AutoCommit on the cloned config before creating the consumer.
	autoCommit bool

	// publishTimeout is reserved for bounded publish calls (use with context in Send/Produce).
	publishTimeout time.Duration

	// writeTimeout is reserved for read/write style timeouts (parity with rabbitmq consume timeout).
	writeTimeout time.Duration

	producerOn bool
	consumerOn bool
}

func withDefaults() *options {
	return &options{
		publishTimeout: 5 * time.Second,
		writeTimeout:   30 * time.Second,
		producerOn:     true,
		consumerOn:     true,
	}
}

// WithProducerOnly initializes the writer only (no single-group Consumer in New).
// You can still use ConsumerManager() for multiple consumers.
func WithProducerOnly() Option {
	return func(o *options) {
		o.producerOn = true
		o.consumerOn = false
	}
}

// WithConsumerOnly initializes the reader only (no Producer in New).
func WithConsumerOnly() Option {
	return func(o *options) {
		o.consumerOn = true
		o.producerOn = false
	}
}

// WithAutoCommit enables automatic offset commits for the in-config consumer (sets Config.Consumer.AutoCommit).
func WithAutoCommit() Option {
	return func(o *options) {
		o.autoCommit = true
	}
}

// WithPublishTimeout sets a suggested timeout for produce operations (stored for callers / future use).
func WithPublishTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout > 0 {
			o.publishTimeout = timeout
		}
	}
}

// WithWriteTimeout sets a suggested timeout for consume/fetch style operations (parity with rabbitmq; stored for future use).
func WithWriteTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout > 0 {
			o.writeTimeout = timeout
		}
	}
}
