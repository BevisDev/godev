package rabbitmq

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Exec interface {
	// Close closes the current channel and connection safely.
	Close()

	// GetConnection returns a live connection, reconnecting if needed.
	// It retries indefinitely until a connection is established.
	GetConnection() (*amqp.Connection, error)

	// GetChannel returns a new channel from the current connection.
	// Returns an error if the connection is not available.
	GetChannel() (*amqp.Channel, error)

	// DeclareQueue is used to create or ensure the existence of a named queue on the RabbitMQ server.
	// If the queue named queueName doesn't exist, it is created
	// If it already exists, this method ensures the existing queue's properties match the declared properties
	// It returns an error if the declaration fails (e.g., due to permission issues or a property mismatch with an existing queue
	DeclareQueue(queueName string) error

	// DeclareQueueWithChannel is used to create or ensure the existence of a named queue on the RabbitMQ server.
	// If the queue named queueName doesn't exist, it is created
	// If it already exists, this method ensures the existing queue's properties match the declared properties
	// It returns an error if the declaration fails (e.g., due to permission issues or a property mismatch with an existing queue
	DeclareQueueWithChannel(channel *amqp.Channel, queueName string) error

	// Publish sends a message to the specified RabbitMQ queue.
	// It supports various message types (string, []byte, numbers, or JSON-serializable objects).
	// Returns an error if the message is too large, or publishing fails
	Publish(ctx context.Context, queueName string, message interface{}) error

	// Consume is used to start receiving messages
	// handler func(...) is a callback function that is executed every time a new message
	// Returns an error if the consumption process cannot be started
	Consume(ctx context.Context, queueName string,
		handler func(ctx context.Context, msg amqp.Delivery)) error
}
