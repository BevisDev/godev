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
	Host       string
	Port       int
	Username   string
	Password   string
	TimeoutSec int
}

type RabbitMQ struct {
	connection *amqp.Connection
	Channel    *amqp.Channel
	TimeoutSec int
}

func NewRabbitMQ(config *RabbitMQConfig) (*RabbitMQ, error) {
	if config == nil {
		return nil, errors.New("config is nil")
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
