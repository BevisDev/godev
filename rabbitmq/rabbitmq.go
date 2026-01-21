package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerHandler func(ctx context.Context, msg Message) error

type ChannelHandler func(ch *amqp.Channel) error

type RabbitMQ struct {
	*Config
	*options
	Queue      *Queue
	Publisher  *Publisher
	connection *amqp.Connection
	mu         sync.RWMutex
}

const (
	// XRid is the header key used to store the RequestID (or trace ID)
	// when publishing, and consumers to retrieve it for logging or tracing.
	XRid = "x-rid"
)

// New creates a new RabbitMQ client using the provided configuration.
//
// It connects to the broker using the AMQP protocol, establishes a connection,
// opens a channel, and returns a `*RabbitMQ` instance.
//
// Returns an error if the configuration is nil, the connection fails,
// or the channel cannot be created.
func New(cf *Config, fs ...OptionFunc) (*RabbitMQ, error) {
	if cf == nil {
		return nil, fmt.Errorf("[rabbitmq] config is nil")
	}

	opt := withDefaults()
	for _, f := range fs {
		f(opt)
	}

	r := &RabbitMQ{
		Config:  cf,
		options: opt,
	}
	conn, err := r.connect()
	if err != nil {
		return nil, err
	}

	r.connection = conn
	r.Queue = newQueue(r)
	r.Publisher = newPublisher(r)
	log.Println("[rabbitmq] connected successfully")
	return r, nil
}

func (r *RabbitMQ) connect() (*amqp.Connection, error) {
	conn, err := amqp.Dial(r.URL())
	if err != nil {
		return nil, err
	}

	if conn == nil {
		return nil, fmt.Errorf("[rabbitmq] connection is nil")
	}

	return conn, nil
}

// Close closes the current channel and connection safely.
func (r *RabbitMQ) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.connection != nil {
		_ = r.connection.Close()
	}
}

// GetConnection returns a live connection, reconnecting if needed.
// It retries indefinitely until a connection is established.
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
		conn, err = r.connect()
		if err == nil {
			log.Println("[rabbitmq] reconnected successfully")
			break
		}

		sleep := time.Second * time.Duration(1<<i)
		log.Printf("[rabbitmq] is attempting to reconnect in %s..., (err: %v)", sleep, err)
		time.Sleep(sleep)
	}
	if conn == nil {
		return nil, err
	}

	r.connection = conn
	return conn, nil
}

// GetChannel returns a new channel from the current connection.
func (r *RabbitMQ) GetChannel() (*amqp.Channel, error) {
	conn, err := r.GetConnection()
	if err != nil {
		return nil, err
	}

	return conn.Channel()
}

func (r *RabbitMQ) WithChannel(fn ChannelHandler) error {
	ch, err := r.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return fn(ch)
}
