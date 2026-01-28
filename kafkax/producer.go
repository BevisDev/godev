package kafkax

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	config *ProducerConfig
	mu     sync.RWMutex
	closed bool
}

func newProducer(cfg *Config) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     cfg.Producer.Balancer,
		BatchSize:    cfg.Producer.BatchSize,
		BatchTimeout: cfg.Producer.BatchTimeout,
		MaxAttempts:  cfg.Producer.MaxAttempts,
		Compression:  cfg.Producer.Compression,
		RequiredAcks: kafka.RequiredAcks(cfg.Producer.RequiredAcks),
		Async:        cfg.Producer.Async,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			fmt.Printf("[kafkax-producer] err: "+msg+"\n", args...)
		}),
	}

	return &Producer{
		writer: writer,
		config: &cfg.Producer,
		closed: false,
	}, nil
}

// Send sends a single message synchronously
func (p *Producer) Send(ctx context.Context, msg *Message) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrProducerClosed
	}

	if msg.Topic == "" {
		return ErrEmptyTopic
	}

	kafkaMsg := kafka.Message{
		Topic: msg.Topic,
		Key:   msg.Key,
		Value: msg.Value,
		Time:  time.Now(),
	}

	// Convert headers
	if len(msg.Headers) > 0 {
		kafkaMsg.Headers = make([]kafka.Header, len(msg.Headers))
		for i, h := range msg.Headers {
			kafkaMsg.Headers[i] = kafka.Header{
				Key:   h.Key,
				Value: h.Value,
			}
		}
	}

	// Set partition if specified
	if msg.Partition >= 0 {
		kafkaMsg.Partition = msg.Partition
	}

	return p.writer.WriteMessages(ctx, kafkaMsg)
}

// SendBatch sends multiple messages in a batch
func (p *Producer) SendBatch(ctx context.Context, messages []*Message) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrProducerClosed
	}

	kafkaMessages := make([]kafka.Message, len(messages))

	for i, msg := range messages {
		if msg.Topic == "" {
			return fmt.Errorf("message %d: %w", i, ErrEmptyTopic)
		}

		kafkaMessages[i] = kafka.Message{
			Topic: msg.Topic,
			Key:   msg.Key,
			Value: msg.Value,
			Time:  time.Now(),
		}

		if len(msg.Headers) > 0 {
			kafkaMessages[i].Headers = make([]kafka.Header, len(msg.Headers))
			for j, h := range msg.Headers {
				kafkaMessages[i].Headers[j] = kafka.Header{
					Key:   h.Key,
					Value: h.Value,
				}
			}
		}

		if msg.Partition >= 0 {
			kafkaMessages[i].Partition = msg.Partition
		}
	}

	return p.writer.WriteMessages(ctx, kafkaMessages...)
}

// SendJSON sends a JSON-encoded message
func (p *Producer) SendJSON(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	return p.Send(ctx, &Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})
}

// SendWithHeaders sends a message with custom headers
func (p *Producer) SendWithHeaders(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	msg := &Message{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	if len(headers) > 0 {
		msg.Headers = make([]Header, 0, len(headers))
		for k, v := range headers {
			msg.Headers = append(msg.Headers, Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	return p.Send(ctx, msg)
}

// Stats returns producer statistics
func (p *Producer) Stats() kafka.WriterStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.writer == nil {
		return kafka.WriterStats{}
	}
	return p.writer.Stats()
}

// Close closes the producer
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// IsClosed returns whether the producer is closed
func (p *Producer) IsClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

func (p *Producer) Produce(
	ctx context.Context,
	topic string,
	key, value []byte,
) error {
	rid := utils.GetRID(ctx)
	var headers []kafka.Header
	headers = append(headers, kafka.Header{
		Key:   consts.XRequestID,
		Value: []byte(rid),
	})

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: headers,
		Time:    time.Now(),
	})
}

func (p *Producer) ProduceBatch(
	ctx context.Context,
	messages []*Message,
) error {
	if len(messages) == 0 {
		return nil
	}

	rid := utils.GetRID(ctx)
	var headers []kafka.Header
	headers = append(headers, kafka.Header{
		Key:   consts.XRequestID,
		Value: []byte(rid),
	})

	msgs := make([]kafka.Message, 0, len(messages))
	for _, msg := range messages {
		kafkaMsg := kafka.Message{
			Topic:   msg.Topic,
			Key:     msg.Key,
			Value:   msg.Value,
			Headers: headers,
			Time:    time.Now(),
		}

		msgs = append(msgs, kafkaMsg)
	}

	return p.writer.WriteMessages(ctx, msgs...)
}
