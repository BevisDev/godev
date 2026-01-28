package kafkax

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	config *ConsumerConfig
	mu     sync.RWMutex
	closed bool
}

// newConsumer creates a new Consumer instance
func newConsumer(cfg *Config) (*Consumer, error) {
	if cfg.Consumer.GroupID == "" {
		return nil, ErrNoGroupID
	}

	if len(cfg.Consumer.Topics) == 0 {
		return nil, ErrNoTopics
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:                cfg.Brokers,
		GroupID:                cfg.Consumer.GroupID,
		GroupTopics:            cfg.Consumer.Topics,
		StartOffset:            cfg.Consumer.StartOffset,
		CommitInterval:         cfg.Consumer.CommitInterval,
		MaxWait:                cfg.Consumer.MaxWait,
		MinBytes:               cfg.Consumer.MinBytes,
		MaxBytes:               cfg.Consumer.MaxBytes,
		PartitionWatchInterval: cfg.Consumer.PartitionWatchInterval,
		SessionTimeout:         cfg.Consumer.SessionTimeout,
		RebalanceTimeout:       cfg.Consumer.RebalanceTimeout,
		HeartbeatInterval:      cfg.Consumer.HeartbeatInterval,
		IsolationLevel:         cfg.Consumer.IsolationLevel,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			fmt.Printf("[kafkax-consumer] err: "+msg+"\n", args...)
		}),
	})

	return &Consumer{
		reader: reader,
		config: &cfg.Consumer,
		closed: false,
	}, nil
}

// Consume starts consuming messages and calls the handler for each message
func (c *Consumer) Consume(ctx context.Context, handler Handler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrConsumerClosed
	}
	c.mu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				log.Printf("[kafkax-consumer] fetching message error: %v\n", err)
				continue
			}

			var rid string
			for _, h := range msg.Headers {
				if consts.XRequestID == h.Key {
					rid = string(h.Value)
				}
			}

			ctxNew := utils.SetValueCtx(nil, consts.RID, rid)
			err = handler(ctxNew, &Message{
				Topic:     msg.Topic,
				Key:       msg.Key,
				Value:     msg.Value,
				Partition: msg.Partition,
				Offset:    msg.Offset,
			})

			// Manual commit after successful processing
			if !c.config.AutoCommit {
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("[kafkax-consumer] error committing message: %v", err)
				}
			}
		}
	}
}

// Stats returns consumer statistics
func (c *Consumer) Stats() kafka.ReaderStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.reader == nil {
		return kafka.ReaderStats{}
	}
	return c.reader.Stats()
}

// Lag returns the current consumer lag
func (c *Consumer) Lag() int64 {
	stats := c.Stats()
	return stats.Lag
}

// SetOffset sets the offset for a specific topic and partition
func (c *Consumer) SetOffset(topic string, partition int, offset int64) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConsumerClosed
	}

	return c.reader.SetOffset(kafka.Partition{
		Topic:     topic,
		Partition: partition,
		Offset:    offset,
	})
}

// Close closes the consumer
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// IsClosed returns whether the consumer is closed
func (c *Consumer) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// convertMessage converts kafka.Message to ConsumedMessage
func (c *Consumer) convertMessage(msg kafka.Message) *ConsumedMessage {
	headers := make(map[string]string)
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	return &ConsumedMessage{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Key:       msg.Key,
		Value:     msg.Value,
		Headers:   headers,
		Time:      msg.Time,
		kafkaMsg:  msg,
		reader:    c.reader,
	}
}
