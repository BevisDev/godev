package kafkax

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Kafka struct {
	cfg      *Config
	producer *Producer
	consumer *Consumer
	mu       sync.RWMutex
	closed   bool
}

// New creates a new Kafka client
// It initializes Producer and/or Consumer based on the config
func New(cfg *Config) (*Kafka, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	k := &Kafka{
		cfg:    cfg,
		closed: false,
	}

	// Initialize producer (always initialized by default)
	k.producer, _ = newProducer(cfg)

	// Initialize consumer only if GroupID and Topics are set
	if cfg.Consumer.GroupID != "" && len(cfg.Consumer.Topics) > 0 {
		consumer, err := newConsumer(cfg)
		if err != nil {
			// Close producer if consumer init fails
			k.producer.Close()
			return nil, fmt.Errorf("failed to create consumer: %w", err)
		}
		k.consumer = consumer
	}

	return k, nil
}

// Producer returns the producer instance
func (k *Kafka) Producer() (*Producer, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.closed {
		return nil, ErrClientClosed
	}

	if k.producer == nil {
		return nil, ErrProducerNotInitialized
	}

	return k.producer, nil
}

// Consumer returns the consumer instance
func (k *Kafka) Consumer() (*Consumer, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.closed {
		return nil, ErrClientClosed
	}

	if k.consumer == nil {
		return nil, ErrConsumerNotInitialized
	}

	return k.consumer, nil
}

// Send is a convenience method to send a message using the producer
func (k *Kafka) Send(ctx context.Context, msg *Message) error {
	producer, err := k.Producer()
	if err != nil {
		return err
	}
	return producer.Send(ctx, msg)
}

// SendJSON is a convenience method to send a JSON message
func (k *Kafka) SendJSON(ctx context.Context, topic string, key string, value interface{}) error {
	producer, err := k.Producer()
	if err != nil {
		return err
	}
	return producer.SendJSON(ctx, topic, key, value)
}

// SendBatch is a convenience method to send multiple messages
func (k *Kafka) SendBatch(ctx context.Context, messages []*Message) error {
	producer, err := k.Producer()
	if err != nil {
		return err
	}
	return producer.SendBatch(ctx, messages)
}

// Consume is a convenience method to consume messages
func (k *Kafka) Consume(ctx context.Context, handler Handler) error {
	consumer, err := k.Consumer()
	if err != nil {
		return err
	}
	return consumer.Consume(ctx, handler)
}

// ConsumeWithRetry is a convenience method to consume with retry logic
func (k *Kafka) ConsumeWithRetry(ctx context.Context, handler Handler, maxRetries int, retryDelay time.Duration) error {
	consumer, err := k.Consumer()
	if err != nil {
		return err
	}

	return consumer.ConsumeWithRetry(ctx, handler, maxRetries, retryDelay)
}

// Close closes both producer and consumer
func (k *Kafka) Close() {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return
	}

	k.closed = true

	if k.producer != nil {
		k.producer.Close()
	}

	if k.consumer != nil {
		k.consumer.Close()
	}
}

// IsClosed returns whether the client is closed
func (k *Kafka) IsClosed() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.closed
}

// HasProducer returns whether producer is initialized
func (k *Kafka) HasProducer() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.producer != nil && !k.producer.IsClosed()
}

// HasConsumer returns whether consumer is initialized
func (k *Kafka) HasConsumer() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.consumer != nil && !k.consumer.IsClosed()
}
