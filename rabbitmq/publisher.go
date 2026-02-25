package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/console"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
	amqp "github.com/rabbitmq/amqp091-go"
)

const maxMessageSize = 50000 // max size of message body in bytes

type Publisher struct {
	mq  *MQ
	log *console.Logger
}

func newPublisher(mq *MQ) *Publisher {
	return &Publisher{
		mq:  mq,
		log: console.New("publisher"),
	}
}

// Send sends a message directly to a single queue (point-to-point).
func (p *Publisher) Send(
	ctx context.Context,
	queueName string,
	message any,
	props ...MsgProperties,
) error {
	return p.publish(ctx, "", queueName, message, props...)
}

// PublishEvent publishes an event to a topic exchange using a routing key.
func (p *Publisher) PublishEvent(
	ctx context.Context,
	exchange string,
	routingKey string,
	message any,
	props ...MsgProperties,
) error {
	return p.publish(ctx, exchange, routingKey, message, props...)
}

// BroadcastEvent publishes an event to all consumers using a fanout exchange.
func (p *Publisher) BroadcastEvent(
	ctx context.Context,
	exchange string,
	message any,
	props ...MsgProperties,
) error {
	return p.publish(ctx, exchange, "", message, props...)
}

// publish is the shared internal publish logic for all publisher APIs.
// It sends a message to the specified exchange and routing key.
// If ctx has no deadline, publishTimeout from MQ options is applied.
func (p *Publisher) publish(
	c context.Context,
	exchange string,
	routingKey string,
	message any,
	props ...MsgProperties,
) error {
	var ctx = c
	if _, ok := c.Deadline(); !ok && p.mq.publishTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.mq.publishTimeout)
		defer cancel()
	}

	return p.mq.WithChannel(func(ch *amqp.Channel) error {
		publishing, err := p.buildPublishing(ctx, message, props...)
		if err != nil {
			return fmt.Errorf("build message: %w", err)
		}
		return ch.PublishWithContext(ctx,
			exchange,
			routingKey,
			true,
			false,
			publishing,
		)
	})
}

func (p *Publisher) buildPublishing(
	ctx context.Context,
	message any,
	props ...MsgProperties,
) (amqp.Publishing, error) {
	var rid = utils.GetRID(ctx)
	prop := new(msgProperties)
	for _, propFn := range props {
		propFn(prop)
	}

	contentType, body, err := p.buildMessage(message)
	if err != nil {
		return amqp.Publishing{}, err
	}

	publishing := amqp.Publishing{
		ContentType: contentType,
		Body:        body,
	}

	if prop.persistentMsg {
		publishing.DeliveryMode = amqp.Persistent
	}

	if !str.IsEmpty(prop.correlationID) {
		publishing.CorrelationId = prop.correlationID
	} else {
		publishing.CorrelationId = rid
	}

	if !str.IsEmpty(prop.messageID) {
		publishing.MessageId = prop.messageID
	}

	if !validate.IsNilOrEmpty(prop.headers) {
		var headers = make(amqp.Table, len(prop.headers))
		for k, v := range prop.headers {
			headers[k] = v
		}
		publishing.Headers = headers
	}

	if !str.IsEmpty(prop.appID) {
		publishing.AppId = prop.appID
	}

	if !str.IsEmpty(prop.replyTo) {
		publishing.ReplyTo = prop.replyTo
	}

	if !str.IsEmpty(prop.userID) {
		publishing.UserId = prop.userID
	}

	if !prop.timestamp.IsZero() {
		publishing.Timestamp = prop.timestamp
	} else {
		publishing.Timestamp = time.Now()
	}

	if !str.IsEmpty(prop.expiration) {
		publishing.Expiration = prop.expiration
	}

	return publishing, nil
}

func (p *Publisher) buildMessage(message any) (string, []byte, error) {
	body, err := utils.ToBytes(message)
	if err != nil {
		return "", nil, err
	}
	if len(body) > maxMessageSize {
		return "", nil, ErrMessageTooLarge
	}
	contentType := consts.TextPlain
	if json.Valid(body) {
		contentType = consts.ApplicationJSON
	}
	return contentType, body, nil
}
