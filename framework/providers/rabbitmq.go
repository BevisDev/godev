package providers

import (
	"context"

	"github.com/BevisDev/godev/rabbitmq"
)

type RabbitMQProvider struct {
	cfg  *rabbitmq.Config
	opts []rabbitmq.Option
	mq   *rabbitmq.MQ
}

func NewRabbitMQProvider(cfg *rabbitmq.Config, opts ...rabbitmq.Option) *RabbitMQProvider {
	return &RabbitMQProvider{
		cfg:  cfg,
		opts: opts,
	}
}

func (p *RabbitMQProvider) Init(ctx context.Context) error {
	mq, err := rabbitmq.New(ctx, p.cfg, p.opts...)
	if err != nil {
		return err
	}
	p.mq = mq
	return nil
}

func (p *RabbitMQProvider) Start(ctx context.Context) error {
	if p.mq != nil && p.mq.Consumer() != nil {
		go p.mq.Consumer().Start(ctx)
	}
	return nil
}

func (p *RabbitMQProvider) Stop(ctx context.Context) error {
	_ = ctx
	if p.mq != nil {
		p.mq.Close()
	}
	return nil
}

func (p *RabbitMQProvider) MQ() *rabbitmq.MQ {
	return p.mq
}
