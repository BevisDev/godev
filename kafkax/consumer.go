package kafkax

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

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
					break
				}
			}

			ctxNew := utils.SetValueCtx(ctx, consts.RID, rid)
			consumed := c.convertMessage(msg)
			err = handler(ctxNew, consumed)

			// Manual commit only after successful processing
			if err != nil {
				log.Printf("[kafkax-consumer] handler error: %v", err)
			} else if !c.config.AutoCommit {
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

	// Reader.SetOffset sets the offset for the current partition; topic/partition
	// parameters are kept for API compatibility but not used here.
	return c.reader.SetOffset(offset)
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

// ReadMessage reads a single message and returns a ConsumedMessage.
func (c *Consumer) ReadMessage(ctx context.Context) (*ConsumedMessage, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, ErrConsumerClosed
	}
	c.mu.RUnlock()

	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}
	return c.convertMessage(msg), nil
}

// CommitMessage commits the offset for the provided ConsumedMessage.
func (c *Consumer) CommitMessage(ctx context.Context, msg *ConsumedMessage) error {
	if msg == nil {
		return nil
	}
	return msg.Commit(ctx)
}

// ConsumeWithRetry wraps Consume with retry logic on handler error.
func (c *Consumer) ConsumeWithRetry(
	ctx context.Context,
	handler Handler,
	maxRetries int,
	retryDelay time.Duration,
) error {
	wrapped := func(ctx context.Context, msg *ConsumedMessage) error {
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			err = handler(ctx, msg)
			if err == nil {
				return nil
			}

			if attempt < maxRetries {
				log.Printf("[kafkax-consumer] handler error: %v, retrying (%d/%d)", err, attempt+1, maxRetries)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryDelay):
				}
			}
		}
		return err
	}

	return c.Consume(ctx, wrapped)
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
