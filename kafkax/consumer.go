package kafkax

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type consumer struct {
	consumer *kafka.Consumer
}

func NewConsumer(cf *Config) (Consumer, error) {
	cfg := cf.ConsumerConfig
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		clientId:         cf.ClientId,
		bootstrapServers: cf.BootstrapServers,
		groupId:          cf.ClientId,
		autoOffsetReset:  cfg.AutoOffsetReset,
		enableAutoCommit: cfg.EnableAutoCommit,
	})
	if err != nil {
		return nil, err
	}

	return &consumer{
		consumer: c,
	}, nil
}

func (c *consumer) Close() {
	if c.consumer != nil {
		_ = c.consumer.Close()
	}
}

func (c *consumer) Subscribe(topics []string) error {
	return c.consumer.SubscribeTopics(topics, nil)
}

func (c *consumer) Consume(timeoutMs int) (*kafka.Message, error) {
	ev := c.consumer.Poll(timeoutMs)
	if ev == nil {
		return nil, nil
	}
	switch e := ev.(type) {
	case *kafka.Message:
		return e, nil
	case kafka.Error:
		return nil, e
	default:
		return nil, fmt.Errorf("unknown event type")
	}
}

func (c *consumer) CommitMessage(msg *kafka.Message) error {
	_, err := c.consumer.CommitMessage(msg)
	return err
}
