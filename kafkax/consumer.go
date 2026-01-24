package kafkax

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Handler interface {
	Handle(msg *kafka.Message) error
}

type Consumer struct {
	consumer   *kafka.Consumer
	autoCommit bool
}

func NewConsumer(cf *Config) (*Consumer, error) {
	cfg := cf.ConsumerConfig
	configMap := kafka.ConfigMap{
		clientId:         cf.ClientId,
		bootstrapServers: cf.BootstrapServers,
	}

	autoCommit := true
	if cfg != nil {
		if cfg.GroupID != "" {
			configMap[groupId] = cfg.GroupID
		} else {
			configMap[groupId] = cf.ClientId
		}
		if cfg.AutoOffsetReset != "" {
			configMap[autoOffsetReset] = cfg.AutoOffsetReset.String()
		}
		autoCommit = cfg.EnableAutoCommit
		configMap[enableAutoCommit] = cfg.EnableAutoCommit
	} else {
		configMap[groupId] = cf.ClientId
		configMap[autoOffsetReset] = Latest.String()
		configMap[enableAutoCommit] = true
	}

	c, err := kafka.NewConsumer(&configMap)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer:   c,
		autoCommit: autoCommit,
	}, nil
}

func (c *Consumer) Close() {
	if c.consumer != nil {
		_ = c.consumer.Close()
	}
}

// Start begins consuming messages from the specified topics and calls the handler for each message.
// It blocks until an error occurs or the Consumer is closed.
// Default timeout is 100ms per poll.
func (c *Consumer) Start(topics []string, handler Handler) error {
	return c.StartWithTimeout(topics, handler, 100)
}

// StartWithTimeout starts consuming messages with a custom timeout per message poll.
// It blocks until an error occurs or the Consumer is closed.
func (c *Consumer) StartWithTimeout(topics []string, handler Handler, timeoutMs int) error {
	if handler == nil {
		return fmt.Errorf("message handler cannot be nil")
	}

	if err := c.consumer.SubscribeTopics(topics, nil); err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	log.Printf("Consumer started, subscribed to topics: %v", topics)

	for {
		ev := c.consumer.Poll(timeoutMs)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			if err := handler.Handle(e); err != nil {
				log.Printf("Error handling message: %v", err)
				// Continue consuming even if handler returns error
				// You can modify this behavior if needed
			} else if !c.autoCommit {
				// Auto-commit is disabled, manually commit after successful processing
				if _, err := c.consumer.CommitMessage(e); err != nil {
					log.Printf("Failed to commit message: %v", err)
				}
			}

		case kafka.Error:
			if e.IsFatal() {
				return fmt.Errorf("fatal Consumer error: %w", e)
			}
			log.Printf("Consumer error: %v", e)

		default:
			log.Printf("Ignored event: %v", e)
		}
	}
}
