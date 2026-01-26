package kafkax

import (
	"context"
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

func (p *Producer) Publish(
	ctx context.Context,
	topic string,
	key, value []byte,
	headers map[string][]byte,
) error {
	var hs []kafka.Header
	for k, v := range headers {
		hs = append(hs, kafka.Header{
			Key:   k,
			Value: v,
		})
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: hs,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
