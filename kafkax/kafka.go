package kafkax

import (
	"context"
	"fmt"
	"sync"

	"github.com/BevisDev/godev/utils/console"
)

type Kafka struct {
	*options
	cfg         *Config
	producer    *Producer
	consumer    *Consumer
	consumerMgr *ConsumerManager
	log         *console.Logger
	mu          sync.RWMutex
	closed      bool
}

// New creates a new Kafka client.
// Use options like WithProducerOnly / WithConsumerOnly / WithAutoCommit (same idea as rabbitmq.New).
// Config is cloned so later changes to cfg do not affect the client.
func New(cfg *Config, opts ...Option) (*Kafka, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	opt := withDefaults()
	for _, f := range opts {
		f(opt)
	}
	if !opt.producerOn && !opt.consumerOn {
		return nil, ErrBothDisabled
	}

	cfg = cfg.clone()
	if opt.autoCommit {
		cfg.Consumer.AutoCommit = true
	}

	if err := cfg.validateBrokers(); err != nil {
		return nil, err
	}
	if opt.producerOn {
		if err := cfg.validateProducerConfig(); err != nil {
			return nil, fmt.Errorf("invalid producer config: %w", err)
		}
	}
	if opt.consumerOn {
		if err := cfg.validateConsumerConfig(); err != nil {
			return nil, fmt.Errorf("invalid consumer config: %w", err)
		}
	}

	k := &Kafka{
		options: opt,
		cfg:     cfg,
		log:     console.New("kafkax"),
		closed:  false,
	}

	if opt.producerOn {
		producer, err := newProducer(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create producer: %w", err)
		}
		k.producer = producer
	}

	if opt.consumerOn {
		consumer, err := newConsumer(cfg)
		if err != nil {
			if k.producer != nil {
				k.producer.Close()
			}
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

// ConsumerManager returns the multi-consumer manager (lazy, same Brokers as this client).
// Use Register / Start to run several consumer groups in one process.
func (k *Kafka) ConsumerManager() (*ConsumerManager, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return nil, ErrClientClosed
	}
	if k.consumerMgr == nil {
		bp := make([]string, len(k.cfg.Brokers))
		copy(bp, k.cfg.Brokers)
		k.consumerMgr = NewConsumerManager(bp)
	}
	return k.consumerMgr, nil
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
func (k *Kafka) SendJSON(ctx context.Context,
	topic string, key string, value interface{},
) error {
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

// Consume is a convenience method to consume messages.
// Retry behavior follows cfg.Consumer.MaxHandlerRetries / HandlerRetryDelay.
func (k *Kafka) Consume(ctx context.Context, handler Handler) error {
	consumer, err := k.Consumer()
	if err != nil {
		return err
	}
	return consumer.Consume(ctx, handler)
}

// Close closes both producer and consumer.
// Logs and ignores close errors so both sides are always attempted.
func (k *Kafka) Close() {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return
	}

	k.closed = true

	if k.producer != nil {
		if err := k.producer.Close(); err != nil {
			k.log.Error("producer close error: %v", err)
		}
		k.producer = nil
	}

	if k.consumer != nil {
		if err := k.consumer.Close(); err != nil {
			k.log.Error("consumer close error: %v", err)
		}
		k.consumer = nil
	}

	if k.consumerMgr != nil {
		k.consumerMgr.Close()
		k.consumerMgr = nil
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
