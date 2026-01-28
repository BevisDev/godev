package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/BevisDev/godev/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerHandler func(ctx context.Context, msg Message) error

type ChannelHandler func(ch *amqp.Channel) error

// RabbitMQ represents a production-ready RabbitMQ client with automatic reconnection
type RabbitMQ struct {
	*options
	config *Config

	connection *amqp.Connection
	connMu     sync.RWMutex

	queue     *Queue
	publisher *Publisher
	consumer  *Consumer

	// Connection lifecycle management
	closeNotify chan *amqp.Error
	reconnectCh chan struct{}

	closed   bool
	closedMu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new RabbitMQ client using the provided configuration.
//
// It connects to the broker using the AMQP protocol, establishes a connection,
// opens a channel, and returns a `*RabbitMQ` instance.
//
// Returns an error if the configuration is nil, the connection fails,
// or the channel cannot be created.
func New(c context.Context, cfg *Config, opts ...Option) (*RabbitMQ, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	opt := withDefaults()
	for _, f := range opts {
		f(opt)
	}

	ctx, cancel := utils.NewCtxCancel(c)

	r := &RabbitMQ{
		config:      cfg.clone(),
		options:     opt,
		reconnectCh: make(chan struct{}, 1),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initial connection
	if err := r.connect(); err != nil {
		cancel()
		return nil, err
	}

	// Initialize components
	r.queue = newQueue(r)
	r.publisher = newPublisher(r)

	// Start connection monitor
	r.wg.Add(1)
	go r.monitorConnection()

	log.Println("[rabbitmq] connected successfully")
	return r, nil
}

// monitorConnection monitors connection health and triggers reconnection
func (r *RabbitMQ) monitorConnection() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			log.Printf("[rabbitmq] context is cannceled")
			return

		case err := <-r.closeNotify:
			if r.isClosed() {
				return
			}

			log.Printf("[rabbitmq] connection closed: %v", err)

			// Trigger reconnection
			select {
			case r.reconnectCh <- struct{}{}:
			default:
			}

			if err := r.reconnect(); err != nil {
				fmt.Printf("[rabbitmq] reconnect failed: %v", err)
			}

		case <-r.reconnectCh:
			if err := r.reconnect(); err != nil {
				fmt.Printf("[rabbitmq] manual reconnection failed: %v", err)
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

	log.Println("[rabbitmq] is shutting down")

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
		log.Println("[rabbitmq] shutdown timeout, forcing close")
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

	log.Println("[rabbitmq] shutdown complete")
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

		log.Printf("[rabbitmq] attempting %d to reconnect...", attempt)

		if err := r.connect(); err != nil {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}

			log.Printf("[rabbitmq] attempting to reconnect failed: %d, err=%v, retry after", delay, err)

			time.Sleep(delay)
			continue
		}

		log.Printf("[rabbitmq] reconnected successfully")
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

	// Trigger reconnection
	select {
	case r.reconnectCh <- struct{}{}:
	default:
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
func (r *RabbitMQ) GetConsumer() *Consumer {
	return r.consumer
}

// GetQueue returns the queue instance
func (r *RabbitMQ) GetQueue() *Queue {
	return r.queue
}
