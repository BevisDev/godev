package framework

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/ginfw/server"
	"github.com/BevisDev/godev/keycloak"
	"github.com/BevisDev/godev/logx"
	"github.com/BevisDev/godev/rabbitmq"
	"github.com/BevisDev/godev/redis"
	"github.com/BevisDev/godev/rest"
	"github.com/BevisDev/godev/scheduler"
)

// Bootstrap manages application lifecycle and dependencies.
type Bootstrap struct {
	// Core services
	Logger    logx.Logger
	Database  *database.Database
	Redis     *redis.Cache
	RabbitMQ  *rabbitmq.RabbitMQ
	Keycloak  keycloak.KeyCloak
	Rest      *rest.Client
	Scheduler *scheduler.Scheduler

	// Configs/options for lazy initialization
	loggerCfg     *logx.Config
	dbCfg         *database.Config
	redisCfg      *redis.Config
	rabbitmqCfg   *rabbitmq.Config
	keycloakCfg   *keycloak.Config
	restOpts      []rest.OptionFunc
	restInit      bool
	schedulerOpt  []scheduler.OptionFunc
	schedulerInit bool
	serverCfg     *server.Config

	// Lifecycle hooks
	beforeInit  []func(ctx context.Context) error
	afterInit   []func(ctx context.Context) error
	beforeStart []func(ctx context.Context) error
	afterStart  []func(ctx context.Context) error
	beforeStop  []func(ctx context.Context) error
	afterStop   []func(ctx context.Context) error

	// Internal state
	mu          sync.RWMutex
	initialized bool
	started     bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// New creates a new Bootstrap instance with the provided options.
func New(opts ...Option) *Bootstrap {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bootstrap{
		ctx:    ctx,
		cancel: cancel,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// BeforeInit registers a hook to run before initialization.
func (b *Bootstrap) BeforeInit(fn func(ctx context.Context) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.beforeInit = append(b.beforeInit, fn)
}

// AfterInit registers a hook to run after initialization.
func (b *Bootstrap) AfterInit(fn func(ctx context.Context) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.afterInit = append(b.afterInit, fn)
}

// BeforeStart registers a hook to run before starting services.
func (b *Bootstrap) BeforeStart(fn func(ctx context.Context) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.beforeStart = append(b.beforeStart, fn)
}

// AfterStart registers a hook to run after starting services.
func (b *Bootstrap) AfterStart(fn func(ctx context.Context) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.afterStart = append(b.afterStart, fn)
}

// BeforeStop registers a hook to run before stopping services.
func (b *Bootstrap) BeforeStop(fn func(ctx context.Context) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.beforeStop = append(b.beforeStop, fn)
}

// AfterStop registers a hook to run after stopping services.
func (b *Bootstrap) AfterStop(fn func(ctx context.Context) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.afterStop = append(b.afterStop, fn)
}

// Init initializes all configured services.
func (b *Bootstrap) Init(ctx context.Context) error {
	b.mu.Lock()
	if b.initialized {
		b.mu.Unlock()
		return errors.New("[bootstrap] already initialized")
	}
	b.mu.Unlock()

	// Run before init hooks
	for _, fn := range b.beforeInit {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("before init hook failed: %w", err)
		}
	}

	log.Println("[bootstrap] initializing services...")

	// Init services in parallel
	errCh := make(chan error, 6)
	var wg sync.WaitGroup

	// Logger
	if b.loggerCfg != nil && b.Logger == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Logger = logx.New(b.loggerCfg)
		}()
	}

	// Database
	if b.dbCfg != nil && b.Database == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := database.New(b.dbCfg)
			if err != nil {
				errCh <- fmt.Errorf("[database] failed to connect: %w", err)
				return
			}
			b.Database = db
		}()
	}

	// Redis
	if b.redisCfg != nil && b.Redis == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache, err := redis.New(b.redisCfg)
			if err != nil {
				errCh <- fmt.Errorf("[redis] failed to connect: %w", err)
				return
			}
			b.Redis = cache
		}()
	}

	// RabbitMQ
	if b.rabbitmqCfg != nil && b.RabbitMQ == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mq, err := rabbitmq.New(b.rabbitmqCfg)
			if err != nil {
				errCh <- fmt.Errorf("init rabbitmq: %w", err)
				return
			}
			b.RabbitMQ = mq
		}()
	}

	// Keycloak
	if b.keycloakCfg != nil && b.Keycloak == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Keycloak = keycloak.New(b.keycloakCfg)
		}()
	}

	// REST client
	if b.Rest == nil && b.restInit {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Rest = rest.New(b.restOpts...)
		}()
	}

	// Scheduler
	if b.Scheduler == nil && b.schedulerInit {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Scheduler = scheduler.New(b.schedulerOpt...)
		}()
	}

	wg.Wait()
	close(errCh)
	if err := <-errCh; err != nil {
		return err
	}

	// Server: create default config if not provided, so server can start without explicit config.
	if b.serverCfg == nil {
		b.serverCfg = &server.Config{}
	}

	// Run after init hooks
	for _, fn := range b.afterInit {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("after init hook failed: %w", err)
		}
	}

	b.mu.Lock()
	b.initialized = true
	b.mu.Unlock()

	log.Println("[bootstrap] initialization completed")
	return nil
}

// Start starts all services and blocks until shutdown signal.
func (b *Bootstrap) Start(ctx context.Context) error {
	b.mu.Lock()
	if !b.initialized {
		b.mu.Unlock()
		return errors.New("[bootstrap] must be initialized before starting")
	}
	if b.started {
		b.mu.Unlock()
		return errors.New("[bootstrap] already started")
	}
	b.mu.Unlock()

	// Run before start hooks
	for _, fn := range b.beforeStart {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("before start hook failed: %w", err)
		}
	}

	log.Println("[bootstrap] starting services...")

	// Start scheduler if configured
	if b.Scheduler != nil {
		b.Scheduler.Start(ctx)
		log.Println("[bootstrap] scheduler started")
	}

	// Start HTTP server if configured
	var serverErrCh chan error
	if b.serverCfg != nil {
		serverErrCh = make(chan error, 1)
		go func() {
			if err := server.Run(ctx, b.serverCfg); err != nil {
				serverErrCh <- err
			}
		}()
	}

	// Run after start hooks
	for _, fn := range b.afterStart {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("after start hook failed: %w", err)
		}
	}

	b.mu.Lock()
	b.started = true
	b.mu.Unlock()

	log.Println("[bootstrap] all services started")

	// Setup signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		log.Println("[bootstrap] context cancelled")
	case s := <-sig:
		log.Printf("[bootstrap] received signal: %v", s)
	case err := <-serverErrCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	return nil
}

// Stop gracefully stops all services.
func (b *Bootstrap) Stop(ctx context.Context) error {
	b.mu.Lock()
	if !b.started {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	log.Println("[bootstrap] stopping services...")

	// Run before stop hooks
	for _, fn := range b.beforeStop {
		if err := fn(ctx); err != nil {
			log.Printf("[bootstrap] before stop hook error: %v", err)
		}
	}

	// Stop scheduler (it stops automatically when context is cancelled)
	if b.Scheduler != nil {
		log.Println("[bootstrap] scheduler stopping...")
	}

	// Close RabbitMQ
	if b.RabbitMQ != nil {
		b.RabbitMQ.Close()
		log.Println("[bootstrap] RabbitMQ closed")
	}

	// Close Redis
	if b.Redis != nil {
		b.Redis.Close()
		log.Println("[bootstrap] Redis closed")
	}

	// Close Database
	if b.Database != nil {
		b.Database.Close()
		log.Println("[bootstrap] database closed")
	}

	// Run after stop hooks
	for _, fn := range b.afterStop {
		if err := fn(ctx); err != nil {
			log.Printf("[bootstrap] after stop hook error: %v", err)
		}
	}

	b.mu.Lock()
	b.started = false
	b.mu.Unlock()

	log.Println("[bootstrap] all services stopped")
	return nil
}

// Run initializes, starts, and manages the application lifecycle.
// It blocks until a shutdown signal is received, then gracefully stops all services.
func (b *Bootstrap) Run(ctx context.Context) error {
	// Initialize
	if err := b.Init(ctx); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Start and wait
	if err := b.Start(ctx); err != nil {
		_ = b.Stop(ctx)
		return fmt.Errorf("start failed: %w", err)
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return b.Stop(shutdownCtx)
}

// Health checks the health of all configured services.
func (b *Bootstrap) Health(ctx context.Context) map[string]error {
	health := make(map[string]error)

	if b.Database != nil {
		if err := b.Database.Ping(); err != nil {
			health["database"] = err
		} else {
			health["database"] = nil
		}
	}

	if b.Redis != nil {
		// Simple ping check
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := b.Redis.Ping(ctxTimeout); err != nil {
			health["redis"] = err
		} else {
			health["redis"] = nil
		}
	}

	if b.RabbitMQ != nil {
		conn, err := b.RabbitMQ.GetConnection()
		if err != nil || conn == nil || conn.IsClosed() {
			health["rabbitmq"] = fmt.Errorf("connection not available")
		} else {
			health["rabbitmq"] = nil
		}
	}

	return health
}

// Context returns the bootstrap context.
func (b *Bootstrap) Context() context.Context {
	return b.ctx
}

// Shutdown triggers graceful shutdown.
func (b *Bootstrap) Shutdown() {
	b.cancel()
}
