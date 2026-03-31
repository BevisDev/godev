package kafkax

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/BevisDev/godev/utils/console"
)

// ManagedConsumer registers one Kafka consumer group + topics + handler.
// IsOn follows rabbitmq.Consumer: set true to start; false skips in Start.
type ManagedConsumer struct {
	Name    string
	IsOn    bool
	Config  ConsumerConfig
	Handler Handler
}

// ConsumerManager runs multiple Kafka consumers (each with its own reader / group / topics).
type ConsumerManager struct {
	brokers []string
	mu      sync.Mutex
	entries map[string]*managedConsumerEntry
	wg      sync.WaitGroup
	log     *console.Logger
}

type managedConsumerEntry struct {
	name    string
	isOn    bool
	cfg     ConsumerConfig
	handler Handler
}

// NewConsumerManager creates a manager for the given broker list (copied).
func NewConsumerManager(brokers []string) *ConsumerManager {
	bp := make([]string, len(brokers))
	copy(bp, brokers)
	return &ConsumerManager{
		brokers: bp,
		entries: make(map[string]*managedConsumerEntry),
		log:     console.New("kafkax-consumer-manager"),
	}
}

// Register adds or replaces a consumer registration.
// Skips when Name is empty, Handler is nil, or Config is invalid.
func (m *ConsumerManager) Register(c *ManagedConsumer) {
	if m == nil || c == nil {
		return
	}
	if c.Name == "" || c.Handler == nil {
		return
	}
	cfg := c.Config
	if len(cfg.Topics) > 0 {
		cfg.Topics = append([]string(nil), cfg.Topics...)
	}
	tmp := &Config{Brokers: m.brokers, Consumer: cfg}
	if err := tmp.validateConsumerConfig(); err != nil {
		m.log.Error("register %s skipped: %v", c.Name, err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.entries[c.Name]; exists {
		m.log.Info("consumer %s already registered, override", c.Name)
	}

	m.entries[c.Name] = &managedConsumerEntry{
		name:    c.Name,
		isOn:    c.IsOn,
		cfg:     cfg,
		handler: c.Handler,
	}
}

// RegisterMany registers multiple consumers.
func (m *ConsumerManager) RegisterMany(list ...*ManagedConsumer) {
	for _, c := range list {
		m.Register(c)
	}
}

// All returns a shallow copy of registrations (Handlers are shared references).
func (m *ConsumerManager) All() map[string]ManagedConsumer {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make(map[string]ManagedConsumer, len(m.entries))
	for k, e := range m.entries {
		cp[k] = ManagedConsumer{
			Name:    e.name,
			IsOn:    e.isOn,
			Config:  e.cfg,
			Handler: e.handler,
		}
	}
	return cp
}

// Close waits for running consumer goroutines (e.g. after ctx cancelled), with timeout.
func (m *ConsumerManager) Close() {
	if m == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		m.log.Info("consumer manager shutdown timeout")
	}
}

// Start launches one goroutine per registered consumer (IsOn == true), then blocks until ctx is done.
func (m *ConsumerManager) Start(ctx context.Context) {
	if m == nil {
		return
	}
	if len(m.brokers) == 0 {
		m.log.Error("no brokers, cannot start consumers")
		return
	}

	list := m.snapshot()
	if len(list) == 0 {
		m.log.Info("no consumer registered")
		return
	}

	m.log.Info("consumer(s) %d registered", len(list))
	started := 0
	for _, e := range list {
		if !e.isOn {
			m.log.Info("consumer %s is off", e.name)
			continue
		}
		started++
		m.wg.Add(1)
		go m.run(ctx, e)
	}
	if started == 0 {
		m.log.Info("no consumer started (all off)")
		return
	}

	m.log.Info("%d consumer(s) started", started)
	<-ctx.Done()

	m.log.Info("shutting down all consumers...")
	m.wg.Wait()
	m.log.Info("all consumers stopped")
}

func (m *ConsumerManager) snapshot() []*managedConsumerEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*managedConsumerEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out
}

func (m *ConsumerManager) run(ctx context.Context, e *managedConsumerEntry) {
	defer m.wg.Done()

	retryDelay := time.Second
	queueName := e.name

	for {
		select {
		case <-ctx.Done():
			return
		default:
			c, err := NewConsumer(m.brokers, e.cfg)
			if err != nil {
				m.log.Error("[%s] create consumer: %v", queueName, err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(retryDelay):
				}
				continue
			}

			err = c.Consume(ctx, e.handler)
			_ = c.Close()

			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				m.log.Error("[%s] consume stopped: %v, retrying...", queueName, err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(retryDelay):
				}
				continue
			}
		}
	}
}
