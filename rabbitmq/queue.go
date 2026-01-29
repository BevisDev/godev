package rabbitmq

import (
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Queue arguments constants
const (
	MessageTTL           = "x-message-ttl"             // Time-to-live for messages (ms)
	DeadLetterExchange   = "x-dead-letter-exchange"    // DLX for failed messages
	DeadLetterRoutingKey = "x-dead-letter-routing-key" // Routing key for DLX
	MaxLength            = "x-max-length"              // Maximum number of messages
	MaxLengthBytes       = "x-max-length-bytes"        // Maximum queue size in bytes
)

// ExchangeType defines how messages are routed from exchange to queues
type ExchangeType string

const (
	// Direct : routes based on exact match of routing key
	// Example: routing key "order.created" only matches binding key "order.created"
	Direct ExchangeType = amqp.ExchangeDirect

	// Topic  : routes based on pattern (* = 1 word, # = 0 or more words)
	// Example: "order.*.email" matches "order.created.email"
	Topic ExchangeType = amqp.ExchangeTopic

	// Fanout : broadcast to all queues (routing key is ignored)
	Fanout ExchangeType = amqp.ExchangeFanout
)

func (e ExchangeType) String() string {
	return string(e)
}

// Queue manages operations related to queue/exchange declarations
type Queue struct {
	mq *MQ
	Spec
}

// QueueSpec defines configuration for a queue
type QueueSpec struct {
	Name       string                 // Queue name (required)
	Durable    bool                   // Survive broker restart (default: true)
	AutoDelete bool                   // Auto-delete when no consumers
	Exclusive  bool                   // Only usable by the connection that created it
	Args       map[string]interface{} // Additional arguments (TTL, DLX, etc.)
}

// ExchangeSpec defines configuration for an exchange
type ExchangeSpec struct {
	Name       string        // Exchange name (required)
	Type       ExchangeType  // Exchange type: Direct, Topic, Fanout
	Durable    bool          // Survive broker restart (default: true)
	AutoDelete bool          // Auto-delete when no bindings
	Internal   bool          // Only other exchanges can publish
	Bindings   []BindingSpec // List of bindings
}

type BindingSpec struct {
	Queue      string
	RoutingKey string
	Args       map[string]interface{}
}

type Spec struct {
	Queues    []QueueSpec
	Exchanges []ExchangeSpec
}

func newQueue(mq *MQ) *Queue {
	return &Queue{
		mq: mq,
	}
}

func (q *Queue) Def(names ...string) error {
	if len(names) == 0 {
		return errors.New("[queue] at least one queue name is required")
	}

	var queues []QueueSpec
	for _, name := range names {
		if name == "" {
			return ErrEmptyQueueName
		}

		queues = append(queues, QueueSpec{
			Name: name,
		})
	}

	return q.Declare(Spec{
		Queues: queues,
	})
}

// Declare declares queues and exchanges according to spec
// Execution order: 1) Queues, 2) Exchanges, 3) Bindings
func (q *Queue) Declare(spec Spec) error {
	q.Spec = spec
	return q.mq.WithChannel(func(ch *amqp.Channel) error {
		// 1. Declare queues first
		if err := q.defQueues(ch, spec.Queues); err != nil {
			return fmt.Errorf("[queue] declare queues: %w", err)
		}

		// 2. Declare exchanges and bindings
		if err := q.defExchanges(ch, spec.Exchanges); err != nil {
			return fmt.Errorf("[queue] declare exchanges: %w", err)
		}

		return nil
	})
}

// defQueues declares all queues in spec
func (q *Queue) defQueues(ch *amqp.Channel, queues []QueueSpec) error {
	for _, qu := range queues {
		if _, err := ch.QueueDeclare(
			qu.Name,
			qu.Durable,    // default: true (survive restart)
			qu.AutoDelete, // default: false
			qu.Exclusive,  // default: false
			false,         // noWait: false
			qu.Args,       // arguments (TTL, DLX, etc.)
		); err != nil {
			return fmt.Errorf("queue '%s': %w", qu.Name, err)
		}
	}
	return nil
}

// defExchanges declares all exchanges and bindings in spec
func (q *Queue) defExchanges(ch *amqp.Channel, exchanges []ExchangeSpec) error {
	for _, ex := range exchanges {
		// Declare exchange
		if err := ch.ExchangeDeclare(
			ex.Name,
			ex.Type.String(),
			ex.Durable,    // default: true (survive restart)
			ex.AutoDelete, // default: false
			ex.Internal,   // default: false
			false,         // noWait: false
			nil,           // args
		); err != nil {
			return fmt.Errorf("exchange '%s': %w", ex.Name, err)
		}

		// Declare bindings
		if err := q.defBindings(ch, ex.Name, ex.Bindings); err != nil {
			return fmt.Errorf("exchange '%s' bindings: %w", ex.Name, err)
		}
	}
	return nil
}

// declareBindings declares all bindings for an exchange
func (q *Queue) defBindings(ch *amqp.Channel, exchangeName string, bindings []BindingSpec) error {
	for _, b := range bindings {
		if err := ch.QueueBind(
			b.Queue,      // queue name
			b.RoutingKey, // routing key
			exchangeName, // exchange name
			false,        // noWait: false
			b.Args,       // args
		); err != nil {
			return fmt.Errorf("bind queue '%s' with key '%s': %w",
				b.Queue, b.RoutingKey, err)
		}
	}
	return nil
}

// Delete deletes a queue
func (q *Queue) Delete(name string, ifUnused, ifEmpty bool) error {
	if name == "" {
		return ErrEmptyQueueName
	}

	return q.mq.WithChannel(func(ch *amqp.Channel) error {
		_, err := ch.QueueDelete(name, ifUnused, ifEmpty, false)
		if err != nil {
			return fmt.Errorf("delete queue '%s': %w", name, err)
		}
		return nil
	})
}

// Purge deletes all messages in a queue
func (q *Queue) Purge(name string) error {
	if name == "" {
		return ErrEmptyQueueName
	}

	return q.mq.WithChannel(func(ch *amqp.Channel) error {
		_, err := ch.QueuePurge(name, false)
		if err != nil {
			return fmt.Errorf("[queue] purge queue '%s': %w", name, err)
		}
		return nil
	})
}
