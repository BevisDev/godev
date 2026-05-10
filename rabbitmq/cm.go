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
	"github.com/BevisDev/godev/utils/str"
	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultConsumerPrefix = "consumer"

// CM is the consumer manager: register Consumer values, then Start(ctx) to run
// queue workers until ctx is cancelled. Optional rabbitmq.WithAutoCommit controls
// whether successful handlers are acked automatically.
type CM struct {
	mq        *MQ
	consumers map[string]*Consumer
	mu        sync.Mutex
	wg        sync.WaitGroup
	log       *console.Logger
}

func newCM(r *MQ) *CM {
	return &CM{
		mq:        r,
		consumers: make(map[string]*Consumer),
		log:       console.New(defaultConsumerPrefix),
	}
}

// Register adds one or more consumers to the manager.
func (m *CM) Register(consumers ...*Consumer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range consumers {
		if c.Handler == nil {
			continue
		}

		q := c.Handler.QueueName()
		if q == "" {
			continue
		}

		if _, ok := m.consumers[q]; ok {
			m.log.Info("queue %s already registered, override", q)
		}

		m.consumers[q] = c
	}
}

func (m *CM) All() map[string]*Consumer {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make(map[string]*Consumer, len(m.consumers))
	for k, v := range m.consumers {
		cp[k] = v
	}

	return cp
}

// Close waits for all consumer goroutines to exit (e.g. after context is cancelled).
func (m *CM) Close() {
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
func (m *CM) Start(ctx context.Context) {
	if len(m.consumers) == 0 {
		m.log.Info("no consumer registered")
		return
	}

	m.log.Info("consumer(s) %d are starting", len(m.consumers))
	for q, c := range m.consumers {
		if !c.IsOn {
			m.log.Info("consumer %s is off", q)
			continue
		}

		m.wg.Add(1)
		go m.Run(ctx, q, c)
	}

	m.log.Info("all consumers started successfully")
	<-ctx.Done()

	m.log.Info("shutting down all consumers...")
	m.wg.Wait()
	m.log.Info("all consumers stopped")
}

// Run runs a single consumer with auto error handling and reconnection.
func (m *CM) Run(ctx context.Context, queueName string, c *Consumer) {
	defer m.wg.Done()

	maxConsecutiveErrors := m.maxConsecutiveErrors
	if c.MaxConsecutiveErrors > 0 {
		maxConsecutiveErrors = c.MaxConsecutiveErrors
	}

	retryDelay := m.retryDelay
	if c.RetryDelay > 0 {
		retryDelay = c.RetryDelay
	}

	errs := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := m.Consume(ctx, queueName, c)
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

// Consume sets up the consumer and processes messages from the queue.
func (m *CM) Consume(ctx context.Context, queueName string, c *Consumer) error {
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

func (m *CM) processMsg(
	queueName string,
	h Handler,
	d amqp.Delivery,
) {
	msg := NewMsgHandler(queueName, d)
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
func (m *CM) handleMsg(ctx context.Context, queueName string, h Handler, msg *MsgHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("[RECOVER][%s] err: %v", queueName, r)
			msg.Reject()
		}
	}()
	return h.Handle(ctx, msg)
}

// newMsgCtx creates a new context with correlation from msg
func (m *CM) newMsgCtx(msg *MsgHandler) context.Context {
	newCtx := utils.NewCtx()
	rid := msg.CorrelationID()
	if str.IsEmpty(rid) {
		rid = utils.GetRID(newCtx)
	}
	newCtx = utils.SetValueCtx(newCtx, consts.RID, rid)

	return newCtx
}

// ConsumeBatch sets up a batching consumer on queueName.
// Each worker collects messages until BatchSize is reached OR FlushInterval
// elapses, then calls BatchConsumer.Handler once and multiple-acks / nacks.
func (m *CM) ConsumeBatch(ctx context.Context, queueName string, c *BatchConsumer) error {
	ch, err := m.mq.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := m.mq.queue.CreateQueues(queueName); err != nil {
		return err
	}

	// ── resolve config ──────────────────────────────────────────────────────
	prefetch := m.prefetchCount
	if c.PrefetchCount > 0 {
		prefetch = c.PrefetchCount
	}

	batchSize := m.batchSize
	if c.BatchSize > 0 {
		batchSize = c.BatchSize
	}

	flushInterval := m.batchFlushInterval
	if c.FlushInterval > 0 {
		flushInterval = c.FlushInterval
	}

	workerCount := m.workerPool
	if c.WorkerPool > 0 {
		workerCount = c.WorkerPool
	}

	// prefetch must cover at least one full batch per worker so workers never
	// stall waiting for messages while the channel window is exhausted.
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}

	// ── start AMQP delivery stream ──────────────────────────────────────────
	msgs, err := ch.ConsumeWithContext(
		ctx,
		queueName,
		"",    // auto-generated consumer tag
		false, // autoAck – we ack manually
		false, false, false, nil,
	)
	if err != nil {
		return err
	}

	// jobs carries individual deliveries to batch workers.
	jobs := make(chan amqp.Delivery, prefetch)

	var workerWG sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			m.batchWorker(ctx, jobs, batchSize, flushInterval, c.Handler)
		}()
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
				return errors.New("rabbitmq: message channel closed")
			}
			select {
			case <-ctx.Done():
				return nil
			case jobs <- d:
			}
		}
	}
}

// batchWorker collects deliveries from jobs, flushes on batchSize or ticker,
// then acks/nacks the entire batch in a single RabbitMQ call.
func (m *CM) batchWorker(
	ctx context.Context,
	jobs <-chan amqp.Delivery,
	batchSize int,
	flushInterval time.Duration,
	handler BatchHandler,
) {
	batch := make([]amqp.Delivery, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		m.processBatch(ctx, handler, batch)
		batch = batch[:0] // reset, keep underlying array
		ticker.Reset(flushInterval)
	}

	for {
		select {
		case <-ctx.Done():
			flush() // drain whatever is left before exit
			return

		case <-ticker.C:
			flush()

		case d, ok := <-jobs:
			if !ok {
				flush() // channel closed – flush remainder
				return
			}
			batch = append(batch, d)
			if len(batch) >= batchSize {
				flush()
			}
		}
	}
}

// processBatch calls the handler and multiple-acks or multiple-nacks the batch.
func (m *CM) processBatch(ctx context.Context, handler BatchHandler, batch []amqp.Delivery) {
	last := batch[len(batch)-1]

	if err := handler(ctx, batch); err != nil {
		// multiple=true nacks everything up to and including last.DeliveryTag
		_ = last.Nack(true /*multiple*/, true /*requeue*/)
		return
	}
	_ = last.Ack(true /*multiple*/)
}
