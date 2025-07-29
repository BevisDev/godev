package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	amqp "github.com/rabbitmq/amqp091-go"
	"strings"
)

type RabbitMQConfig struct {
	Host     string // RabbitMQ server host
	Port     int    // RabbitMQ server port
	Username string // Username for authentication
	Password string // Password for authentication
	VHost    string // VHost Virtual host
}

type RabbitMQ struct {
	config     *RabbitMQConfig
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

const (
	// maxMessageSize max size message
	maxMessageSize = 50000

	// xstate is the header key used to store the RequestID (or trace ID)
	// when publishing, and consumers to retrieve it for logging or tracing.
	xstate = "x-state"
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

	conn, err := amqp.Dial(
		fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
			config.Username, config.Password, config.Host, config.Port, config.VHost),
	)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		config:     config,
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		_ = r.Channel.Close()
	}
	if r.Connection != nil {
		_ = r.Connection.Close()
	}
}

func (r *RabbitMQ) DeclareQueue(queueName string) error {
	_, err := r.Channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

// Publish sends a message to the specified RabbitMQ queue.
// It supports various message types (string, []byte, numbers, or JSON-serializable objects).
// Returns an error if the message is too large, or publishing fails
func (r *RabbitMQ) Publish(ctx context.Context, queueName string, message interface{}) error {
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

	// Assume queue is already declared during initialization
	// If needed, add a check for queue existence instead of declaring it every time

	var state = utils.GetState(ctx)
	return r.Channel.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
			Headers: amqp.Table{
				"x-state": state,
			},
		},
	)
}

func (r *RabbitMQ) Consume(ctx context.Context, queueName string,
	handler func(ctx context.Context, msg amqp.Delivery)) error {
	msgs, err := r.Channel.ConsumeWithContext(ctx,
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
		newCtx := utils.NewCtx(nil)
		if raw, ok := msg.Headers[xstate]; ok {
			if s, ok := raw.(string); ok {
				newCtx = utils.NewCtxWithState(newCtx, s)
			}
		}
		handler(newCtx, msg)
	}

	return nil
}
