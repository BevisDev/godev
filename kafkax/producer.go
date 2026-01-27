package kafkax

import (
	"context"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string,

// opts ...ProducerOption,
) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
	}

	//for _, opt := range opts {
	//	opt(w)
	//}

	return &Producer{writer: w}
}

func (p *Producer) Close() {
	if p.writer != nil {
		_ = p.writer.Close()
	}
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
