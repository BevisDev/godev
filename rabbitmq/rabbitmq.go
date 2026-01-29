package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/console"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelHandler func(ch *amqp.Channel) error

// RabbitMQ represents a production-ready RabbitMQ client with automatic reconnection
type RabbitMQ struct {
	*options
	config *Config

	connection *amqp.Connection
	connMu     sync.RWMutex

	queue     *Queue
	publisher *Publisher
	consumer  *ConsumerManager

	// Connection lifecycle management
	closeNotify chan *amqp.Error
	reconnectCh chan struct{}

	closed   bool
	closedMu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// logger
	log *console.Logger
}

// New creates a new RabbitMQ client using the provided configuration.
//
// It connects to the broker using the AMQP protocol, establishes a connection,
// opens a channel, and returns a `*RabbitMQ` instance.
//
// Returns an error if the configuration is nil, the connection fails,
// or the channel cannot be created.
func New(cfg *Config, opts ...Option) (*RabbitMQ, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	opt := withDefaults()
	for _, f := range opts {
		f(opt)
	}

	ctx, cancel := utils.NewCtxCancel(context.Background())

	r := &RabbitMQ{
		config:      cfg.clone(),
		options:     opt,
		reconnectCh: make(chan struct{}, 1),
		ctx:         ctx,
		cancel:      cancel,
		log:         console.New("rabbitmq"),
	}

	// Initial connection
	if err := r.connect(); err != nil {
		cancel()
		return nil, err
	}

	// Initialize components
	r.queue = newQueue(r)
	if r.publisherOn {
		r.publisher = newPublisher(r)
	}

	if r.consumerOn {
		r.consumer = newConsumer(r)
	}

	// Start connection monitor
	r.wg.Add(1)
	go r.monitorConnection()

	r.log.Info("connected successfully")
	return r, nil
}

// monitorConnection monitors connection health and triggers reconnection
func (r *RabbitMQ) monitorConnection() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			r.log.Info("context is cancelled")
			return

		case err := <-r.closeNotify:
			if r.isClosed() {
				return
			}

			r.log.Info("connection closed: %v", err)

			// Trigger reconnection
			select {
			case r.reconnectCh <- struct{}{}:
			default:
			}

			if err := r.reconnect(); err != nil {
				r.log.Error("reconnect failed: %v", err)
			}

		case <-r.reconnectCh:
			if err := r.reconnect(); err != nil {
				r.log.Error("manual reconnection failed: %v", err)
			}
		}
	}
}

func (r *RabbitMQ) connect() error {
	r.connMu.Lock()
	defer r.connMu.Unlock()

	// Close existing connection if any
	if r.connection != nil && !r.connection.IsClosed() {
		_ = r.connection.Close()
	}

	conn, err := amqp.Dial(r.config.URL())
	if err != nil {
		return fmt.Errorf("[rabbitmq] failed to dial: %w", err)
	}

	if conn == nil {
		return errors.New("[rabbitmq] connection is nil after dial")
	}

	r.connection = conn
	r.closeNotify = make(chan *amqp.Error, 1)
	r.connection.NotifyClose(r.closeNotify)

	return nil
}

// isClosed checks if the client is closed
func (r *RabbitMQ) isClosed() bool {
	r.closedMu.RLock()
	defer r.closedMu.RUnlock()
	return r.closed
}

// Close gracefully shuts down the RabbitMQ client
func (r *RabbitMQ) Close() {
	r.closedMu.Lock()
	if r.closed {
		r.closedMu.Unlock()
		return
	}

	r.closed = true
	r.closedMu.Unlock()

	r.log.Info("shutting down")

	// Cancel context to stop background goroutines
	r.cancel()

	// Wait for background goroutines
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		r.log.Info("shutdown timeout, forcing close")
	}

	// Close components
	//if r.consumer != nil {
	//	r.consumer.Close()
	//}

	//if r.publisher != nil {
	//	r.publisher.
	//}

	// Close connection
	r.connMu.Lock()
	if r.connection != nil {
		_ = r.connection.Close()
	}
	r.connMu.Unlock()

	r.log.Info("shutdown complete")
}

// reconnect attempts to reconnect with exponential backoff
func (r *RabbitMQ) reconnect() error {
	if r.isClosed() {
		return ErrClientClosed
	}

	maxRetries := 10
	baseDelay := time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		default:
		}

		r.log.Info("attempting %d to reconnect...", attempt)

		if err := r.connect(); err != nil {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}

			r.log.Info("attempting to reconnect failed: %d, err=%v, retry after", delay, err)

			time.Sleep(delay)
			continue
		}

		r.log.Info("reconnected successfully")
		return nil
	}

	return ErrMaxRetriesReached
}

// Health checks the health status of the connection
func (r *RabbitMQ) Health() error {
	if r.isClosed() {
		return ErrClientClosed
	}

	return r.WithChannel(func(ch *amqp.Channel) error {
		// Try to declare a temporary queue to verify channel works
		_, err := ch.QueueDeclare(
			"",    // name (auto-generated)
			false, // durable
			true,  // delete when unused
			true,  // exclusive
			false, // no-wait
			nil,   // arguments
		)
		return err
	})
}

// GetConnection returns a live connection, reconnecting if needed.
// It retries indefinitely until a connection is established.
func (r *RabbitMQ) GetConnection() (*amqp.Connection, error) {
	if r.isClosed() {
		return nil, ErrClientClosed
	}

	r.connMu.RLock()
	conn := r.connection
	r.connMu.RUnlock()

	// return connection if not closed
	if conn != nil && !conn.IsClosed() {
		return conn, nil
	}

	// attempt to reconnect using the existing reconnect logic
	if err := r.reconnect(); err != nil {
		return nil, err
	}

	r.connMu.RLock()
	defer r.connMu.RUnlock()

	if r.connection == nil || r.connection.IsClosed() {
		return nil, errors.New("[rabbitmq] connection is nil or closed after reconnect")
	}

	return r.connection, nil
}

// GetChannel returns a new channel from the current connection.
func (r *RabbitMQ) GetChannel() (*amqp.Channel, error) {
	conn, err := r.GetConnection()
	if err != nil {
		return nil, err
	}

	return conn.Channel()
}

// GetConsumerChannel returns a channel WITH QoS configured for consuming
// Use this specifically for consumers
func (r *RabbitMQ) GetConsumerChannel() (*amqp.Channel, error) {
	conn, err := r.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	// Set QoS for consumer
	if r.options.prefetchCount > 0 {
		if err := ch.Qos(
			r.options.prefetchCount, // prefetch count
			0,                       // prefetch size (0 = no limit)
			false,                   // global (false = per consumer)
		); err != nil {
			ch.Close()
			return nil, fmt.Errorf("set qos: %w", err)
		}
	}

	return ch, nil
}

func (r *RabbitMQ) WithChannel(fn ChannelHandler) error {
	ch, err := r.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return fn(ch)
}

// GetPublisher returns the publisher instance
func (r *RabbitMQ) GetPublisher() *Publisher {
	return r.publisher
}

// GetConsumer returns the consumer instance
func (r *RabbitMQ) GetConsumer() *ConsumerManager {
	return r.consumer
}

// GetQueue returns the queue instance
func (r *RabbitMQ) GetQueue() *Queue {
	return r.queue
}
