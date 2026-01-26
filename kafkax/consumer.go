package kafkax

import (
	"context"
	"errors"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/segmentio/kafka-go"
	"log"
)

type Consumer struct {
	reader *kafka.Reader
	cf     *Config
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

func (c *Consumer) Close() {
	if c.reader != nil {
		_ = c.reader.Close()
	}
}

func (c *Consumer) Consume(
	ctx context.Context,
	topics []string,
	handler Handler,
) error {
	if len(topics) == 0 {
		return errors.New("[kafkax] consume non topic")
	}

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

				log.Printf("[kafkax] consume message error: %v\n", err)
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

			if err == nil {
				_ = c.reader.CommitMessages(ctx, msg)
			}
		}
	}
}
