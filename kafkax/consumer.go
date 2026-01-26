package kafkax

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(
	brokers []string,
	groupID string,
	topic string,
// opts ...ConsumerOption,
) *Consumer {

	cfg := kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,

		MinBytes: 1e3,
		MaxBytes: 10e6,
	}

	//for _, opt := range opts {
	//	opt(&cfg)
	//}

	return &Consumer{
		reader: kafka.NewReader(cfg),
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Run(ctx context.Context, handler Handler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}

			headers := map[string][]byte{}
			for _, h := range msg.Headers {
				headers[h.Key] = h.Value
			}

			err = handler(ctx, Message{
				Topic:     msg.Topic,
				Key:       msg.Key,
				Value:     msg.Value,
				Headers:   headers,
				Partition: msg.Partition,
				Offset:    msg.Offset,
			})

			if err == nil {
				_ = c.reader.CommitMessages(ctx, msg)
			}
		}
	}
}
