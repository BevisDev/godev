package rabbitmq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/console"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Handler defines the interface for message consumers.
type Handler interface {
	// Handle processes a single message. Returns nil to ack, error to requeue.
	Handle(ctx context.Context, msg Message) error
}

type Consumer struct {
	IsOn    bool   // enable / disable consumer
	Queue   string // queue name
	Handler Handler
}

// ConsumerManager manages multiple consumers with auto-reconnect and error handling.
type ConsumerManager struct {
	mq        *RabbitMQ
	consumers map[string]*Consumer
	mu        sync.Mutex
	wg        sync.WaitGroup
	log       *console.Logger
}

// newConsumer creates a new ConsumerManager for the given RabbitMQ instance.
func newConsumer(r *RabbitMQ) *ConsumerManager {
	return &ConsumerManager{
		mq:        r,
		consumers: make(map[string]*Consumer),
		log:       console.New("rabbitmq-consumer"),
	}
}

// Register adds one or more consumers to the manager.
func (m *ConsumerManager) Register(consumers ...*Consumer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range consumers {
		if c.Queue == "" || c.Handler == nil {
			continue
		}

		if _, ok := m.consumers[c.Queue]; ok {
			m.log.Info("queue %s already registered, override", c.Queue)
		}

		m.consumers[c.Queue] = c
	}
}

func (m *ConsumerManager) All() map[string]*Consumer {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make(map[string]*Consumer, len(m.consumers))
	for k, v := range m.consumers {
		cp[k] = v
	}

	return cp
}

// Start starts all registered consumers in separate goroutines until context is canceled.
func (m *ConsumerManager) Start(ctx context.Context) {
	if len(m.consumers) == 0 {
		m.log.Info("no consumer registered")
		return
	}

	m.log.Info("consumer(s) %d are starting", len(m.consumers))
	for _, consumer := range m.consumers {
		m.wg.Add(1)
		go m.run(ctx, consumer)
	}

	m.log.Info("all consumers started successfully")
	<-ctx.Done()

	m.log.Info("shutting down all consumers...")
	m.wg.Wait()
	m.log.Info("all consumers stopped")
}

// run runs a single consumer with auto error handling and reconnection.
func (m *ConsumerManager) run(ctx context.Context, consumer *Consumer) {
	defer m.wg.Done()

	queueName := consumer.Queue
	consecutiveErrors := 0
	const maxConsecutiveErrors = 10
	const retryDelay = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := m.consume(ctx, consumer)

			if err != nil {
				consecutiveErrors++
				m.log.Info("[%s] error: %v (consecutive errors: %d)",
					queueName, err, consecutiveErrors)

				if consecutiveErrors >= maxConsecutiveErrors {
					m.log.Info("[%s] exceeded max consecutive errors (%d), stopping",
						queueName, maxConsecutiveErrors)
					return
				}

				var amqpErr *amqp.Error
				if errors.As(err, &amqpErr) {
					if amqpErr.Code == 504 || amqpErr.Code == 320 || amqpErr.Code == 501 {
						m.log.Warn("[%s] connection error, reconnecting...", queueName)
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
func (m *ConsumerManager) consume(ctx context.Context, consumer Handler) error {
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

	m.log.Info("[%s] started consuming messages", queueName)

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
				m.log.Info("[rabbitmq] consumer [%s] handle error: %v", queueName, err)
				msg.Requeue()
			} else {
				msg.Commit()
			}
		}
	}
}

// createMessageContext creates a new context with x-rid from message headers.
func (m *ConsumerManager) createMessageContext(msg Message) context.Context {
	newCtx := utils.NewCtx()
	if xRID := msg.Header(consts.XRequestID); xRID != nil {
		if s, ok := xRID.(string); ok && s != "" {
			newCtx = utils.SetValueCtx(newCtx, consts.RID, s)
		}
	}

	return newCtx
}
