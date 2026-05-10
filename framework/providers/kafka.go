package providers

import (
	"context"
	"time"

	"github.com/BevisDev/godev/kafkax"
)

type KafkaProvider struct {
	cfg        *kafkax.Config
	handler    kafkax.Handler
	withRetry  bool
	maxRetries int
	retryDelay time.Duration
	kafka      *kafkax.Kafka
}

func NewKafkaProvider(cfg *kafkax.Config, handler kafkax.Handler) *KafkaProvider {
	return &KafkaProvider{
		cfg:     cfg,
		handler: handler,
	}
}

func (p *KafkaProvider) WithRetry(maxRetries int, retryDelay time.Duration) *KafkaProvider {
	p.withRetry = true
	p.maxRetries = maxRetries
	p.retryDelay = retryDelay
	return p
}

func (p *KafkaProvider) Init(ctx context.Context) error {
	_ = ctx
	k, err := kafkax.New(p.cfg)
	if err != nil {
		return err
	}
	p.kafka = k
	return nil
}

func (p *KafkaProvider) Start(ctx context.Context) error {
	if p.kafka == nil || !p.kafka.HasConsumer() || p.handler == nil {
		return nil
	}

	if p.withRetry {
		go func() {
			_ = p.kafka.ConsumeWithRetry(ctx, p.handler, p.maxRetries, p.retryDelay)
		}()
		return nil
	}

	go func() {
		_ = p.kafka.Consume(ctx, p.handler)
	}()
	return nil
}

func (p *KafkaProvider) Stop(ctx context.Context) error {
	_ = ctx
	if p.kafka != nil {
		p.kafka.Close()
	}
	return nil
}

func (p *KafkaProvider) Kafka() *kafkax.Kafka {
	return p.kafka
}
