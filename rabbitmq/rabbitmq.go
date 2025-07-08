package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils/jsonx"

	"github.com/BevisDev/godev/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	Host       string // RabbitMQ server host
	Port       int    // RabbitMQ server port
	Username   string // Username for authentication
	Password   string // Password for authentication
	TimeoutSec int    // Timeout in seconds for message operations
}

type RabbitMQ struct {
	connection *amqp.Connection
	Channel    *amqp.Channel
	TimeoutSec int
}

// defaultTimeoutSec defines the default timeout (in seconds) for rabbitmq operations.
const defaultTimeoutSec = 10

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

func (r *RabbitMQ) Publish(c context.Context, queueName string, message interface{}) error {
	json := jsonx.ToJSONBytes(message)
	if len(json) > 50000 {
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

func (r *RabbitMQ) Subscribe(queueName string, handler func(amqp.Delivery)) error {
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
		for msg := range msgs {
			handler(msg)
		}
	}()

	return nil
}
