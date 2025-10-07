package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strings"
	"sync"
	"time"
)

type RabbitMQ struct {
	*Config
	connection *amqp.Connection
	mu         sync.RWMutex
}

const (
	// maxMessageSize max size message
	maxMessageSize = 50000

	// Xstate is the header key used to store the RequestID (or trace ID)
	// when publishing, and consumers to retrieve it for logging or tracing.
	Xstate = "x-state"
)

// New creates a new RabbitMQ client using the provided configuration.
//
// It connects to the broker using the AMQP protocol, establishes a connection,
// opens a channel, and returns a `*RabbitMQ` instance.
//
// Returns an error if the configuration is nil, the connection fails,
// or the channel cannot be created.
func New(cf *Config) (Exec, error) {
	if cf == nil {
		return nil, errors.New("config is nil")
	}

	r := &RabbitMQ{
		Config: cf,
	}

	conn, err := r.init()
	if err != nil {
		return nil, err
	}

	r.connection = conn
	return r, nil
}

func (r *RabbitMQ) init() (*amqp.Connection, error) {
	conn, err := amqp.Dial(
		fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
			r.Username, r.Password, r.Host, r.Port, r.VHost),
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (r *RabbitMQ) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.connection != nil {
		_ = r.connection.Close()
		r.connection = nil
	}
}

func (r *RabbitMQ) GetConnection() (*amqp.Connection, error) {
	r.mu.RLock()

	// return connection if not closed
	if r.connection != nil && !r.connection.IsClosed() {
		defer r.mu.RUnlock()
		return r.connection, nil
	}

	r.mu.RUnlock()

	// reconnect only one go routine
	r.mu.Lock()
	defer r.mu.Unlock()

	var (
		conn *amqp.Connection
		err  error
	)
	for i := 0; i < 5; i++ {
		conn, err = r.init()
		if err == nil {
			log.Println("reconnect RabbitMQ success")
			break
		}

		sleep := time.Second * time.Duration(1<<i)
		log.Printf("failed to reconnect RabbitMQ: %v, retrying in %s...", err, sleep)
		time.Sleep(sleep)
	}
	if conn == nil {
		return nil, err
	}

	r.connection = conn
	return conn, nil
}

func (r *RabbitMQ) GetChannel() (*amqp.Channel, error) {
	conn, err := r.GetConnection()
	if err != nil {
		return nil, err
	}

	return conn.Channel()
}

func (r *RabbitMQ) DeclareQueue(queueName string) error {
	ch, err := r.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := r.DeclareQueueWithChannel(ch, queueName); err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) DeclareQueueWithChannel(channel *amqp.Channel, queueName string) error {
	_, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declare queue %s, err = %w", queueName, err)
	}
	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, queueName string, message interface{}) error {
	ch, err := r.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	var (
		body        []byte
		contentType string
	)
	switch v := message.(type) {
	case []byte:
		body = v
		contentType = consts.TextPlain

	case string:
		body = []byte(v)
		trimmed := strings.TrimSpace(v)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			contentType = consts.ApplicationJSON
		} else {
			contentType = consts.TextPlain
		}

	case int, int64, float64, bool:
		body = []byte(fmt.Sprint(v))
		contentType = consts.TextPlain

	default:
		body = jsonx.ToJSONBytes(v)
		contentType = consts.ApplicationJSON
	}
	if len(body) > maxMessageSize {
		return fmt.Errorf("message is too large: %d", len(body))
	}

	if err := r.DeclareQueueWithChannel(ch, queueName); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	var state = utils.GetState(ctx)
	return ch.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
			Headers: amqp.Table{
				Xstate: state,
			},
		},
	)
}

func (r *RabbitMQ) Consume(ctx context.Context, queueName string,
	handler func(ctx context.Context, msg amqp.Delivery)) error {
	ch, err := r.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer ch.Close()

	if err := r.DeclareQueueWithChannel(ch, queueName); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := ch.ConsumeWithContext(ctx,
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		newCtx := utils.NewCtx()
		if raw, ok := msg.Headers[Xstate]; ok {
			if s, ok := raw.(string); ok {
				newCtx = utils.SetValueCtx(newCtx, consts.State, s)
			}
		}
		handler(newCtx, msg)
	}

	return nil
}
