package kafkax

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/console"
	"github.com/segmentio/kafka-go"
)

const (
	defaultFetchRetryDelay           = 300 * time.Millisecond
	defaultHandlerRetryDelay         = 500 * time.Millisecond
	defaultMaxConsecutiveFetchErrors = 20
	defaultWorkerPool                = 10
)

type Consumer struct {
	reader *kafka.Reader
	config *ConsumerConfig
	log    *console.Logger
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

	lg := console.New("kafkax-consumer")
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
			lg.Error(msg, args...)
		}),
	})

	return &Consumer{
		reader: reader,
		config: &cfg.Consumer,
		log:    lg,
		closed: false,
	}, nil
}

// NewConsumer builds a consumer from brokers and consumer settings.
// Brokers must be non-empty; ConsumerConfig must pass the same validation as Config.Consumer.
func NewConsumer(brokers []string, cc ConsumerConfig) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, ErrNoBrokers
	}
	cfg := &Config{
		Brokers:  brokers,
		Consumer: cc,
	}
	if err := cfg.validateConsumerConfig(); err != nil {
		return nil, err
	}
	return newConsumer(cfg)
}

// Consume starts consuming messages and calls the handler for each message.
// If ConsumerConfig.MaxHandlerRetries > 0, handler failures are retried before poison handling.
func (c *Consumer) Consume(ctx context.Context, handler Handler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrConsumerClosed
	}
	c.mu.RUnlock()

	handler = c.wrapHandlerWithRetries(handler)

	workerCount := c.config.WorkerPool
	if workerCount <= 0 {
		workerCount = defaultWorkerPool
	}

	jobs := make(chan kafka.Message, workerCount)
	var workerWG sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWG.Go(func() {
			for msg := range jobs {
				c.processMsg(ctx, handler, msg)
			}
		})
	}
	defer func() {
		close(jobs)
		workerWG.Wait()
	}()

	consecutiveFetchErrors := 0
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

				consecutiveFetchErrors++
				c.log.Error(
					"fetching message error: %v (consecutive=%d)",
					err, consecutiveFetchErrors,
				)
				if consecutiveFetchErrors >= defaultMaxConsecutiveFetchErrors {
					return fmt.Errorf(
						"[kafkax-consumer] exceeded max consecutive fetch errors (%d): %w",
						defaultMaxConsecutiveFetchErrors, err,
					)
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(defaultFetchRetryDelay):
				}
				continue
			}
			consecutiveFetchErrors = 0
			select {
			case <-ctx.Done():
				return ctx.Err()
			case jobs <- msg:
			}
		}
	}
}

func (c *Consumer) processMsg(ctx context.Context, handler Handler, msg kafka.Message) {
	msgCtx := c.newMsgCtx(ctx, msg)
	consumed := c.convertMessage(msg)

	if err := c.handleMsg(msgCtx, handler, consumed); err != nil {
		c.log.Error("handler error: %v", err)
		return
	}

	if !c.config.AutoCommit {
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("error committing message: %v", err)
		}
	}
}

func (c *Consumer) handleMsg(
	ctx context.Context,
	handler Handler,
	msg *ConsumedMessage,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[kafkax-consumer] recovered panic: %v", r)
		}
	}()
	return handler(ctx, msg)
}

func (c *Consumer) newMsgCtx(ctx context.Context, msg kafka.Message) context.Context {
	rid := utils.GetRID(ctx)
	for _, h := range msg.Headers {
		if consts.XRequestID == h.Key {
			rid = string(h.Value)
			break
		}
	}
	return utils.SetValueCtx(ctx, consts.RID, rid)
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

func (c *Consumer) wrapHandlerWithRetries(handler Handler) Handler {
	maxR := c.config.MaxHandlerRetries
	if maxR <= 0 {
		return handler
	}
	delay := c.config.HandlerRetryDelay
	if delay <= 0 {
		delay = defaultHandlerRetryDelay
	}
	return func(ctx context.Context, msg *ConsumedMessage) error {
		var err error
		for attempt := 0; attempt <= maxR; attempt++ {
			err = handler(ctx, msg)
			if err == nil {
				return nil
			}
			if attempt < maxR {
				c.log.Warn("handler error: %v, retrying (%d/%d)", err, attempt+1, maxR)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
			}
		}
		c.log.Error("retries exhausted for topic=%s partition=%d offset=%d: %v (message committed/skipped)",
			msg.Topic, msg.Partition, msg.Offset, err)
		if !c.config.AutoCommit {
			_ = msg.Commit(ctx)
		}
		return nil
	}
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
