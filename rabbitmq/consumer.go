package rabbitmq

import (
	"context"
	"errors"
	"fmt"
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
	// When autoCommit is false, handler must call msg.Commit() on success.
	Handle(ctx context.Context, msg *MsgHandler) error
}

type Consumer struct {
	IsOn    bool   // enable / disable consumer
	Queue   string // queue name
	Handler Handler
}

// ConsumerManager manages multiple consumers with auto-reconnect and error handling.
type ConsumerManager struct {
	mq        *MQ
	consumers map[string]*Consumer
	mu        sync.Mutex
	wg        sync.WaitGroup
	log       *console.Logger
}

// newConsumer creates a new ConsumerManager for the given MQ instance.
func newConsumer(r *MQ) *ConsumerManager {
	return &ConsumerManager{
		mq:        r,
		consumers: make(map[string]*Consumer),
		log:       console.New("consumer"),
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

// Close waits for all consumer goroutines to exit (e.g. after context is cancelled).
func (m *ConsumerManager) Close() {
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		m.log.Info("consumer shutdown timeout")
	}
}

// Start starts all registered consumers in separate goroutines until context is canceled.
func (m *ConsumerManager) Start(ctx context.Context) {
	if len(m.consumers) == 0 {
		m.log.Info("no consumer registered")
		return
	}

	m.log.Info("consumer(s) %d are starting", len(m.consumers))
	for _, consumer := range m.consumers {
		if !consumer.IsOn {
			m.log.Info("consumer %s is off", consumer.Queue)
			continue
		}

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
				m.log.Error("[%s] error: %v (consecutive errors: %d)",
					queueName, err, consecutiveErrors)

				if consecutiveErrors >= maxConsecutiveErrors {
					m.log.Error("[%s] exceeded max consecutive errors (%d), stopping",
						queueName, maxConsecutiveErrors)
					return
				}

				var amqpErr *amqp.Error
				if errors.As(err, &amqpErr) {
					if amqpErr.Code == 504 || amqpErr.Code == 320 || amqpErr.Code == 501 {
						m.log.Error("[%s] connection error, reconnecting...", queueName)
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
func (m *ConsumerManager) consume(ctx context.Context, consumer *Consumer) error {
	queueName := consumer.Queue

	ch, err := m.mq.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := m.mq.queue.CreateQueues(queueName); err != nil {
		return err
	}

	if err := ch.Qos(m.mq.prefetchCount, 0, false); err != nil {
		return err
	}

	msgs, err := ch.ConsumeWithContext(
		ctx,
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

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed")
			}

			msg := &MsgHandler{
				queueName: queueName,
				d:         d,
			}
			msgCtx := m.NewMsgCtx(msg)

			if err := m.handleMsg(msgCtx, queueName, consumer.Handler, msg); err != nil {
				m.log.Info("[%s] error: %v", queueName, err)
				if !m.mq.autoCommit {
					msg.Requeue()
				}
				continue
			}

			if m.mq.autoCommit {
				msg.Commit()
			}
		}
	}
}

// handleMsg runs Handler.Handle and recovers from panic.
func (m *ConsumerManager) handleMsg(
	ctx context.Context,
	queueName string,
	h Handler,
	msg *MsgHandler,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[RECOVER][%s] err: %v", queueName, r)
			msg.Reject()
		}
	}()
	return h.Handle(ctx, msg)
}

// NewMsgCtx creates a new context with correlation from msg
func (m *ConsumerManager) NewMsgCtx(msg *MsgHandler) context.Context {
	newCtx := utils.NewCtx()
	newCtx = utils.SetValueCtx(newCtx,
		consts.RID,
		msg.CorrelationID(),
	)

	return newCtx
}
