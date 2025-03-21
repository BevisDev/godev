package rabbitmq

import (
	"fmt"

	"github.com/BevisDev/godev/helper"
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
	channel    *amqp.Channel
	timeoutSec int
}

func NewRabbitMQ(config *RabbitMQConfig) (*RabbitMQ, error) {
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
		channel:    ch,
		timeoutSec: config.TimeoutSec,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.connection != nil {
		r.connection.Close()
	}
}

func (r *RabbitMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) PutMessageToQueue(queueName string, message interface{}) error {
	json := helper.ToJSON(message)
	if len(json) > 50000 {
		return fmt.Errorf("Message is too large: %d", len(json))
	}
	ctx, cancel := helper.CreateCtxTimeout(nil, r.timeoutSec)
	defer cancel()

	q, err := r.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	err = r.channel.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: helper.ApplicationJSON,
			Body:        json,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) ConsumeMessage(queueName string, handler func(amqp.Delivery)) error {
	q, err := r.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	msgs, err := r.channel.Consume(
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
