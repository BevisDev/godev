package rabbitmq

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Handler defines the interface for message consumers.
type Handler interface {
	// Queue returns the queue name to consume from.
	Queue() string

	// Handle processes a single message. Returns nil to ack, error to requeue.
	Handle(ctx context.Context, msg Message) error
}

// Consumer manages multiple consumers with auto-reconnect and error handling.
type Consumer struct {
	mq        *RabbitMQ
	consumers []Handler
	wg        sync.WaitGroup
}

// Register creates a new Consumer for the given RabbitMQ instance.
func Register(r *RabbitMQ) *Consumer {
	return &Consumer{
		mq:        r,
		consumers: make([]Handler, 0),
	}
}

// Register adds one or more consumers to the manager.
func (m *Consumer) Register(consumers ...Handler) {
	m.consumers = append(m.consumers, consumers...)
}

// Start starts all registered consumers in separate goroutines until context is canceled.
func (m *Consumer) Start(ctx context.Context) {
	if len(m.consumers) == 0 {
		log.Println("[rabbitmq] no consumers registered")
		return
	}

	log.Printf("[rabbitmq] consumer(s) %d are starting", len(m.consumers))
	for _, consumer := range m.consumers {
		m.wg.Add(1)
		go m.run(ctx, consumer)
	}

	log.Println("[rabbitmq] all consumers started successfully")
	<-ctx.Done()

	log.Println("[rabbitmq] shutting down all consumers...")
	m.wg.Wait()
	log.Println("[rabbitmq] all consumers stopped")
}

// run runs a single consumer with auto error handling and reconnection.
func (m *Consumer) run(ctx context.Context, consumer Handler) {
	defer m.wg.Done()

	queueName := consumer.Queue()
	consecutiveErrors := 0
	const maxConsecutiveErrors = 10
	const retryDelay = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Printf("[rabbitmq] consumer [%s] stopped by context cancellation", queueName)
			return
		default:
			err := m.consume(ctx, consumer)

			if err != nil {
				consecutiveErrors++
				log.Printf("[rabbitmq] consumer [%s] error: %v (consecutive errors: %d)",
					queueName, err, consecutiveErrors)

				if consecutiveErrors >= maxConsecutiveErrors {
					log.Printf("[rabbitmq] consumer [%s] exceeded max consecutive errors (%d), stopping",
						queueName, maxConsecutiveErrors)
					return
				}

				var amqpErr *amqp.Error
				if errors.As(err, &amqpErr) {
					if amqpErr.Code == 504 || amqpErr.Code == 320 || amqpErr.Code == 501 {
						log.Printf("[WARNING][rabbitmq] consumer [%s] connection error, reconnecting...", queueName)
						time.Sleep(retryDelay)
						continue
					}
				}
				time.Sleep(retryDelay)
			} else {
				consecutiveErrors = 0
			}
		}
	}
}

// consume sets up the consumer and processes messages from the queue.
func (m *Consumer) consume(ctx context.Context, consumer Handler) error {
	queueName := consumer.Queue()

	ch, err := m.mq.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := m.mq.Queue.DeclareSimple(queueName); err != nil {
		return err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := ch.ConsumeWithContext(
		ctx,
		queueName,
		"",    // consumer tag (empty = auto-generated)
		false, // auto-ack (false = manual ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	log.Printf("[rabbitmq] consumer [%s] started consuming messages", queueName)

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed")
			}

			msg := Message{Delivery: delivery}
			msgCtx := m.createMessageContext(msg)

			if err := consumer.Handle(msgCtx, msg); err != nil {
				log.Printf("[rabbitmq] consumer [%s] handle error: %v", queueName, err)
				msg.Requeue()
			} else {
				msg.Commit()
			}
		}
	}
}

// createMessageContext creates a new context with x-rid from message headers.
func (m *Consumer) createMessageContext(msg Message) context.Context {
	newCtx := utils.NewCtx()
	if xRID := msg.Header(consts.XRequestID); xRID != nil {
		if s, ok := xRID.(string); ok && s != "" {
			newCtx = utils.SetValueCtx(newCtx, consts.RID, s)
		}
	}

	return newCtx
}
