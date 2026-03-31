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
	IsOn bool // enable / disable consumer

	Queue string // queue name

	Handler Handler

	// PrefetchCount sets the AMQP QoS prefetch count (messages) per queue/consumer.
	// If <= 0, it falls back to the MQ default (WithPrefetchCount).
	PrefetchCount int

	// WorkerPool is number of concurrent workers for this consumer.
	// If <= 0, it falls back to 10
	WorkerPool int

	// MaxConsecutiveErrors caps the number of consecutive consume errors before stopping this consumer.
	// If <= 0, it falls back to 10.
	MaxConsecutiveErrors int

	// RetryDelay is the delay between retries after a consume error.
	// If <= 0, it falls back to 5 seconds.
	RetryDelay time.Duration
}

// ConsumerManager manages multiple consumers with auto-reconnect and error handling.
type ConsumerManager struct {
	mq                   *MQ
	consumers            map[string]*Consumer
	mu                   sync.Mutex
	wg                   sync.WaitGroup
	log                  *console.Logger
	maxConsecutiveErrors int
	retryDelay           time.Duration
	prefetchCount        int
	workerPool           int
}

// newConsumer creates a new ConsumerManager for the given MQ instance.
func newConsumer(r *MQ) *ConsumerManager {
	return &ConsumerManager{
		mq:                   r,
		consumers:            make(map[string]*Consumer),
		log:                  console.New("consumer"),
		maxConsecutiveErrors: 10,
		retryDelay:           5 * time.Second,
		prefetchCount:        1,
		workerPool:           10,
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
func (m *ConsumerManager) run(ctx context.Context, c *Consumer) {
	defer m.wg.Done()

	maxConsecutiveErrors := m.maxConsecutiveErrors
	if c.MaxConsecutiveErrors > 0 {
		maxConsecutiveErrors = c.MaxConsecutiveErrors
	}

	retryDelay := m.retryDelay
	if c.RetryDelay > 0 {
		retryDelay = c.RetryDelay
	}

	queueName := c.Queue
	errs := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := m.consume(ctx, c)
			if err != nil {
				errs++
				m.log.Error("[%s] error: %v (consecutive errors: %d)",
					queueName, err, errs)

				if errs >= maxConsecutiveErrors {
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
				errs = 0
			}
		}
	}
}

// consume sets up the consumer and processes messages from the queue.
func (m *ConsumerManager) consume(ctx context.Context, c *Consumer) error {
	queueName := c.Queue

	ch, err := m.mq.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := m.mq.queue.CreateQueues(queueName); err != nil {
		return err
	}

	prefetch := m.prefetchCount
	if c.PrefetchCount > 0 {
		prefetch = c.PrefetchCount
	}

	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}

	workerCount := m.workerPool
	if c.WorkerPool > 0 {
		workerCount = c.WorkerPool
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

	jobs := make(chan amqp.Delivery, workerCount)

	var workerWG sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWG.Go(func() {
			for d := range jobs {
				m.processMsg(queueName, c.Handler, d)
			}
		})
	}

	defer func() {
		close(jobs)
		workerWG.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil

		case d, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed")
			}
			select {
			case <-ctx.Done():
				return nil
			case jobs <- d:
			}
		}
	}
}

func (m *ConsumerManager) processMsg(
	queueName string,
	h Handler,
	d amqp.Delivery,
) {
	msg := &MsgHandler{
		queueName: queueName,
		d:         d,
	}
	msgCtx := m.newMsgCtx(msg)

	if err := m.handleMsg(msgCtx, queueName, h, msg); err != nil {
		m.log.Info("[%s] error: %v", queueName, err)
		if !m.mq.autoCommit {
			msg.Requeue()
		}
		return
	}

	if m.mq.autoCommit {
		m.log.Info("[%s] committed correlationID: %s",
			queueName, msg.CorrelationID())
		msg.Commit()
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

// newMsgCtx creates a new context with correlation from msg
func (m *ConsumerManager) newMsgCtx(msg *MsgHandler) context.Context {
	newCtx := utils.NewCtx()
	newCtx = utils.SetValueCtx(newCtx,
		consts.RID,
		msg.CorrelationID(),
	)

	return newCtx
}
