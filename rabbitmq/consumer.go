package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer interface {
	// Queue returns the queue name to consume from
	Queue() string

	// Handle processes a single message
	Handle(ctx context.Context, d amqp.Delivery) error

	// OnError is called when Handle returns an error (optional hook)
	OnError(ctx context.Context, d amqp.Delivery, err error)
}

type ConsumerManager struct {
	*RabbitMQ
	consumers []Consumer
	wg        sync.WaitGroup
}

func Register(r *RabbitMQ) *ConsumerManager {
	return &ConsumerManager{
		RabbitMQ:  r,
		consumers: make([]Consumer, 0),
	}
}

func (m *ConsumerManager) Register(consumers ...Consumer) {
	m.consumers = append(m.consumers, consumers...)
}

// Start starts and manages all defined message consumers in separate goroutines.
//
// It launches each consumer defined in the `consumersList` and monitors them in a loop.
// If a consumer exits with an error or panics, it will automatically be restarted
// after a short delay. If the main context is canceled, all consumers are stopped gracefully.
//
// Each consumer function must match the signature `func(ctx context.Context) error`.
func (m *ConsumerManager) Start(ctx context.Context) {
	if len(m.consumers) == 0 {
		log.Println("No consumers registered")
	}

	log.Printf("Starting %d consumer(s)\n", len(m.consumers))

	for _, consumer := range m.consumers {
		m.wg.Add(1)
		c := consumer
		fmt.Print(c)
		//go m.Run(ctx, c)
	}
	log.Println("Consumers started successfully")

	// wait for context cancellation
	<-ctx.Done()

	log.Println("Shutting down all consumers...")
	m.wg.Wait()
	log.Println("All consumers stopped")
}

//
//func (m *ConsumerManager) Run(ctx context.Context, consumer Consumer) {
//	defer m.wg.Done()
//
//	queue := consumer.Queue()
//
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		default:
//			err := m.handleConsumer(ctx, consumer)
//
//			if err != nil {
//				consecutiveErrors++
//				m.logger.Error(ctx, "Consumer [%s] error: %v (consecutive errors: %d)",
//					queue, err, consecutiveErrors)
//
//				// Check max retries
//				if m.config.MaxRetries > 0 && consecutiveErrors >= m.config.MaxRetries {
//					m.logger.Error(ctx, "Consumer [%s] exceeded max retries (%d), stopping",
//						queue, m.config.MaxRetries)
//					return
//				}
//
//				// Handle AMQP connection errors
//				var amqpErr *amqp091.Error
//				if errors.As(err, &amqpErr) {
//					if amqpErr.Code == 504 || amqpErr.Code == 320 {
//						m.logger.Warn(ctx, "Consumer [%s] connection error, reconnecting...", queue)
//						if reconnectErr := m.client.Reconnect(); reconnectErr != nil {
//							m.logger.Error(ctx, "Failed to reconnect: %v", reconnectErr)
//						}
//						time.Sleep(m.config.ReconnectDelay)
//						continue
//					}
//				}
//
//				time.Sleep(m.config.RetryDelay)
//			} else {
//				// Reset counter on successful run
//				consecutiveErrors = 0
//			}
//		}
//	}
//}
//
//func (m *ConsumerManager) Ha(ctx context.Context, consumer Consumer) {
//	defer m.wg.Done()
//
//	queue := consumer.Queue()
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		default:
//			err := m.handleConsumer(ctx, consumer)
//
//			if err != nil {
//				consecutiveErrors++
//				m.logger.Error(ctx, "Consumer [%s] error: %v (consecutive errors: %d)",
//					queue, err, consecutiveErrors)
//
//				// Check max retries
//				if m.config.MaxRetries > 0 && consecutiveErrors >= m.config.MaxRetries {
//					m.logger.Error(ctx, "Consumer [%s] exceeded max retries (%d), stopping",
//						queue, m.config.MaxRetries)
//					return
//				}
//
//				// Handle AMQP connection errors
//				var amqpErr *amqp091.Error
//				if errors.As(err, &amqpErr) {
//					if amqpErr.Code == 504 || amqpErr.Code == 320 {
//						m.logger.Warn(ctx, "Consumer [%s] connection error, reconnecting...", queue)
//						if reconnectErr := m.client.Reconnect(); reconnectErr != nil {
//							m.logger.Error(ctx, "Failed to reconnect: %v", reconnectErr)
//						}
//						time.Sleep(m.config.ReconnectDelay)
//						continue
//					}
//				}
//
//				time.Sleep(m.config.RetryDelay)
//			} else {
//				// Reset counter on successful run
//				consecutiveErrors = 0
//			}
//		}
//	}
//}
