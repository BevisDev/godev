package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ExchangeType defines how messages are routed from an exchange to queues in RabbitMQ.
//
// In general:
//   - Direct  : route by exact routing key match
//   - Topic   : route by routing key pattern (most common for events)
//   - Fanout  : broadcast to all bound queues (routing key is ignored)
type ExchangeType string

const (
	// Direct exchange
	//
	// Routes messages to queues whose binding key EXACTLY matches
	// the message routing key.
	//
	// Example:
	//   routing key: "order.created"
	//
	//   Bindings:
	//     order.created.queue -> "order.created"   (MATCH)
	//     order.*.queue       -> (patterns are NOT supported) (NOT MATCH)
	Direct ExchangeType = amqp.ExchangeDirect

	// Topic exchange
	//
	// Routes messages based on routing key PATTERNS.
	// This is the most commonly used exchange type in event-driven systems.
	//
	// Pattern symbols:
	//   * : matches exactly one word
	//   # : matches zero or more words
	//
	// Example:
	//   routing key: "order.created.email"
	//
	//   Bindings:
	//     order.*.email  -> MATCH
	//     order.#        -> MATCH
	//     user.*         -> NOT MATCH
	Topic ExchangeType = amqp.ExchangeTopic

	// Fanout exchange
	//
	// Broadcasts messages to ALL queues bound to the exchange.
	// The routing key is completely ignored.
	//
	// Example:
	//   A single published message is delivered to every bound queue.
	Fanout ExchangeType = amqp.ExchangeFanout
)

func (e ExchangeType) String() string {
	return string(e)
}

const (
	MessageTTL           = "x-message-ttl"
	DeadLetterExchange   = "x-dead-letter-exchange"
	DeadLetterRoutingKey = "x-dead-letter-routing-key"
)

type Queue struct {
	mq *RabbitMQ
	Spec
}

type QueueSpec struct {
	Queue string
}

type ExchangeSpec struct {
	Name     string
	Type     ExchangeType
	Bindings []BindingSpec
}

type BindingSpec struct {
	Queue      string
	RoutingKey string
}

type Spec struct {
	Queues    []QueueSpec
	Exchanges []ExchangeSpec
}

func newQueue(mq *RabbitMQ) *Queue {
	return &Queue{
		mq: mq,
	}
}

func (q *Queue) DeclareSimple(names ...string) error {
	var queues []QueueSpec
	for _, name := range names {
		queues = append(queues, QueueSpec{
			Queue: name,
		})
	}

	return q.Declare(Spec{
		Queues: queues,
	})
}

func (q *Queue) Declare(spec Spec) error {
	q.Spec = spec
	return q.mq.WithChannel(func(ch *amqp.Channel) error {
		// 1. declare queues
		for _, qu := range spec.Queues {
			_, err := ch.QueueDeclare(
				qu.Queue,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				return fmt.Errorf("error declare queue %s, err = %w", qu.Queue, err)
			}
		}

		// 2. declare exchanges + bindings
		for _, ex := range spec.Exchanges {
			if err := ch.ExchangeDeclare(
				ex.Name,
				ex.Type.String(),
				true,
				false,
				false,
				false,
				nil,
			); err != nil {
				return err
			}

			// declare bindings
			for _, b := range ex.Bindings {
				if err := ch.QueueBind(
					b.Queue,
					b.RoutingKey,
					ex.Name,
					false,
					nil,
				); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
