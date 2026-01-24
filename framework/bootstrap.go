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
	"github.com/BevisDev/godev/migration"
	"github.com/BevisDev/godev/rabbitmq"
	"github.com/BevisDev/godev/redis"
	"github.com/BevisDev/godev/rest"
	"github.com/BevisDev/godev/scheduler"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

// Bootstrap manages application lifecycle and dependencies.
type Bootstrap struct {
	*options

	// Core services
	Logger    logx.Logger
	Database  *database.Database
	Migration migration.Migration
	Redis     *redis.Cache
	RabbitMQ  *rabbitmq.RabbitMQ
	Keycloak  *keycloak.KeyCloak
	Rest      *rest.Client
	Scheduler *scheduler.Scheduler
	HTTPApp   *server.HTTPApp

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
func New(opts ...OptionFunc) *Bootstrap {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bootstrap{
		options: new(options),
		ctx:     ctx,
		cancel:  cancel,
	}

	for _, opt := range opts {
		opt(b.options)
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
			return fmt.Errorf("[bootstrap] before init failed: %w", err)
		}
	}

	log.Println("[bootstrap] initializing services...")

	// Logger: must be initialized first (synchronously) for other services
	if b.loggerConf != nil && b.Logger == nil {
		b.Logger = logx.New(b.loggerConf)
	}

	// Init services in parallel (except logger which must be first)
	g, ctx := errgroup.WithContext(ctx)

	// Database
	if b.dbConf != nil && b.Database == nil {
		g.Go(func() error {
			db, err := database.New(b.dbConf)
			if err != nil {
				return fmt.Errorf("[database] %w", err)
			}
			b.Database = db
			return nil
		})
	}

	// Redis
	if b.redisConf != nil && b.Redis == nil {
		g.Go(func() error {
			cache, err := redis.New(b.redisConf)
			if err != nil {
				return fmt.Errorf("[redis] %w", err)
			}
			b.Redis = cache
			return nil
		})
	}

	// RabbitMQ
	if b.rabbitmqConf != nil && b.RabbitMQ == nil {
		g.Go(func() error {
			mq, err := rabbitmq.New(b.rabbitmqConf)
			if err != nil {
				return fmt.Errorf("[rabbitmq] %w", err)
			}
			b.RabbitMQ = mq
			return nil
		})
	}

	// Keycloak
	if b.keycloakConf != nil && b.Keycloak == nil {
		g.Go(func() error {
			b.Keycloak = keycloak.New(b.keycloakConf)
			return nil
		})
	}

	// Scheduler
	if b.schedulerOn && b.Scheduler == nil {
		g.Go(func() error {
			b.Scheduler = scheduler.New(b.schedulerOpt...)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// REST client: init after logger is ready (may need logger)
	// If logger is not in options, inject it automatically
	if b.restOn && b.Rest == nil {
		opts := b.restOpts
		// Check if WithLogger is already in options by checking if logger was passed
		// Since we can't easily check, we'll always inject logger if available
		// (rest.New will handle duplicates gracefully or user can avoid passing nil)
		if b.Logger != nil {
			// Inject logger - if user passed nil, this will override it
			opts = append(opts, rest.WithLogger(b.Logger))
		}
		b.Rest = rest.New(opts...)
	}

	// Run after init hooks (services are now available, can set Setup/Shutdown here)
	for _, fn := range b.afterInit {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("[bootstrap] after init hook failed: %w", err)
		}
	}

	// Ensure server config exists (for setting Setup/Shutdown later)
	if b.serverConf == nil {
		b.serverConf = &server.Config{
			Shutdown: func(ctx context.Context) error {
				if b.Logger != nil {
					b.Logger.Sync()
					b.Logger = nil
				}
				if b.Database != nil {
					b.Database.Close()
					b.Database = nil
				}
				if b.Redis != nil {
					b.Redis.Close()
					b.Redis = nil
				}
				if b.RabbitMQ != nil {
					b.RabbitMQ.Close()
					b.RabbitMQ = nil
				}
				return nil
			},
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
			return fmt.Errorf("[bootstrap] before start hook failed: %w", err)
		}
	}

	log.Println("[bootstrap] starting services...")

	// Start scheduler if configured
	if b.Scheduler != nil {
		b.Scheduler.Start(ctx)
	}

	// Start HTTP server if configured
	if b.serverConf != nil {
		b.HTTPApp = server.New(b.serverConf)
		if err := b.HTTPApp.Start(); err != nil {
			return fmt.Errorf("[bootstrap] failed to start HTTP server: %w", err)
		}
		// Server errors are handled internally by HTTPApp
		// We don't need to monitor errCh separately since Start() is non-blocking
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

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		log.Println("[bootstrap] root context cancelled")
	case s := <-sig:
		log.Printf("[bootstrap] received signal: %v", s)
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

	// Stop HTTP server if configured
	if b.HTTPApp != nil {
		if err := b.HTTPApp.Stop(ctx); err != nil {
			log.Printf("[bootstrap] HTTP server stop error: %v", err)
		} else {
			log.Println("[bootstrap] HTTP server stopped")
		}
	}

	if b.Logger != nil {
		b.Logger.Sync()
		b.Logger = nil
	}

	// Close RabbitMQ
	if b.RabbitMQ != nil {
		b.RabbitMQ.Close()
		b.RabbitMQ = nil
	}

	// Close Redis
	if b.Redis != nil {
		b.Redis.Close()
		b.Redis = nil
	}

	// Close Database
	if b.Database != nil {
		b.Database.Close()
		b.Database = nil
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
func (b *Bootstrap) Health(ctx context.Context) map[string]interface{} {
	health := make(map[string]interface{})

	if b.Database != nil {
		if err := b.Database.Ping(); err != nil {
			health["database"] = err
		} else {
			health["database"] = "OK"
		}
	}

	if b.Redis != nil {
		// Simple ping check
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := b.Redis.Ping(ctxTimeout); err != nil {
			health["redis"] = err
		} else {
			health["redis"] = "OK"
		}
	}

	if b.RabbitMQ != nil {
		conn, err := b.RabbitMQ.GetConnection()
		if err != nil || conn == nil || conn.IsClosed() {
			health["rabbitmq"] = fmt.Errorf("connection not available")
		} else {
			health["rabbitmq"] = "OK"
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

// SetServerSetup sets the server Setup function after services are initialized.
// This allows Setup to access initialized services (Logger, Database, Redis, etc.).
// Should be called in AfterInit hook or after Init() completes.
func (b *Bootstrap) SetServerSetup(setup func(r *gin.Engine)) {
	if b.serverConf == nil {
		b.serverConf = &server.Config{}
	}
	b.serverConf.Setup = setup
}

// SetServerShutdown sets the server Shutdown function after services are initialized.
// This allows Shutdown to access initialized services for cleanup.
// Should be called in AfterInit hook or after Init() completes.
func (b *Bootstrap) SetServerShutdown(shutdown func(ctx context.Context) error) {
	if b.serverConf == nil {
		b.serverConf = &server.Config{}
	}
	b.serverConf.Shutdown = shutdown
}
