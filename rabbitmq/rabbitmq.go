package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	Host       string // RabbitMQ server host
	Port       int    // RabbitMQ server port
	Username   string // Username for authentication
	Password   string // Password for authentication
	TimeoutSec int    // TimeoutSec in seconds for message operations
}

type RabbitMQ struct {
	connection *amqp.Connection
	Channel    *amqp.Channel
	TimeoutSec int
}

const (
	// defaultTimeoutSec defines the default timeout (in seconds) for rabbitmq operations.
	defaultTimeoutSec = 10

	// maxMessageSize max size message
	maxMessageSize = 50000
)

// NewRabbitMQ creates a new RabbitMQ client using the provided configuration.
//
// It connects to the broker using the AMQP protocol, establishes a connection,
// opens a channel, and returns a `*RabbitMQ` instance.
//
// Returns an error if the configuration is nil, the connection fails,
// or the channel cannot be created.
//
// Example:
//
//	cfg := &RabbitMQConfig{
//	    Host:     "localhost",
//	    Port:     5672,
//	    Username: "guest",
//	    Password: "guest",
//	}
//
//	client, err := NewRabbitMQ(cfg)
//	if err != nil {
//	    log.Fatalf("failed to connect to RabbitMQ: %v", err)
//	}
func NewRabbitMQ(config *RabbitMQConfig) (*RabbitMQ, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.TimeoutSec == 0 {
		config.TimeoutSec = defaultTimeoutSec
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		config.Username, config.Password, config.Host, config.Port))
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &RabbitMQ{
		connection: conn,
		Channel:    ch,
		TimeoutSec: config.TimeoutSec,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}
}

func (r *RabbitMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) Ack(d amqp.Delivery) error {
	return d.Ack(false)
}

func (r *RabbitMQ) Nack(d amqp.Delivery, requeue bool) error {
	return d.Nack(false, requeue)
}

func (r *RabbitMQ) Publish(c context.Context, queueName string, message interface{}) error {
	if err := c.Err(); err != nil {
		return err
	}

	json := jsonx.ToJSONBytes(message)
	if len(json) > maxMessageSize {
		return fmt.Errorf("message is too large: %d", len(json))
	}

	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	q, err := r.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	err = r.Channel.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: consts.ApplicationJSON,
			Body:        json,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) Subscribe(ctx context.Context, queueName string, handler func(amqp.Delivery)) error {
	q, err := r.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	msgs, err := r.Channel.Consume(
		q.Name,
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

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				handler(msg)
			}
		}
	}()

	return nil
}
