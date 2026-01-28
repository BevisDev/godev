package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/console"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/str"
	amqp "github.com/rabbitmq/amqp091-go"
)

const maxMessageSize = 50000 // maxMessageSize max size message

type Publisher struct {
	mq  *RabbitMQ
	log *console.Logger
}

func newPublisher(mq *RabbitMQ) *Publisher {
	return &Publisher{
		mq:  mq,
		log: console.New("rabbitmq-publisher"),
	}
}

// Send sends a message directly to a single queue (point-to-point).
func (p *Publisher) Send(ctx context.Context, queueName string, message interface{}) error {
	return p.publish(ctx, "", queueName, message)
}

// PublishEvent publishes an event to a topic exchange using a routing key.
func (p *Publisher) PublishEvent(ctx context.Context, exchange, routingKey string, message interface{}) error {
	return p.publish(ctx, exchange, routingKey, message)
}

// BroadcastEvent publishes an event to all consumers using a fanout exchange.
func (p *Publisher) BroadcastEvent(ctx context.Context, exchange string, message interface{}) error {
	return p.publish(ctx, exchange, "", message)
}

// publish is the shared internal publish logic for all publisher APIs.
// It sends a message to the specified exchange and routing key.
func (p *Publisher) publish(ctx context.Context,
	exchange, routingKey string,
	message interface{},
) error {
	return p.mq.WithChannel(func(ch *amqp.Channel) error {
		contentType, body, err := p.buildMessage(message)
		if err != nil {
			return err
		}

		publishing := amqp.Publishing{
			ContentType: contentType,
			Body:        body,
			Headers: amqp.Table{
				consts.XRequestID: utils.GetRID(ctx),
			},
		}
		if p.mq.persistentMsg {
			publishing.DeliveryMode = amqp.Persistent
		}

		return ch.PublishWithContext(ctx,
			exchange,
			routingKey,
			false,
			false,
			publishing,
		)
	})
}

func (p *Publisher) buildMessage(message interface{}) (string, []byte, error) {
	var (
		body        []byte
		contentType = consts.TextPlain
	)
	switch v := message.(type) {
	case []byte:
		body = v
		if json.Valid(v) {
			contentType = consts.ApplicationJSON
			break
		}
	case string:
		body = []byte(v)
		if json.Valid(body) {
			contentType = consts.ApplicationJSON
		}
	case bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		body = []byte(str.ToString(v))
	default:
		var err error
		body, err = jsonx.ToJSONBytes(v)
		if err != nil {
			return "", nil, err
		}
		contentType = consts.ApplicationJSON
	}
	if len(body) > maxMessageSize {
		return "", nil, fmt.Errorf("message is too large: %d", len(body))
	}

	return contentType, body, nil
}
