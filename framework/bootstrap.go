package framework

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BevisDev/godev/kafkax"
	"github.com/BevisDev/godev/utils/console"

	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/ginfw/server"
	"github.com/BevisDev/godev/keycloak"
	"github.com/BevisDev/godev/logger"
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
	log *console.Logger

	// Core services
	logger     *logger.Logger
	database   *database.DB
	migration  *migration.Migration
	redisCache *redis.Cache
	rabbitmq   *rabbitmq.MQ
	keycloak   *keycloak.KC
	kafka      *kafkax.Kafka
	restClient *rest.Client
	scheduler  *scheduler.Scheduler

	// server
	httpApp *server.HTTPApp

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
		options: new(options),
		log:     console.New("bootstrap"),
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
		return errors.New("[boostrap] already initialized")
	}
	b.mu.Unlock()

	// Consume before init hooks
	for _, fn := range b.beforeInit {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("[bootstrap] before init failed: %w", err)
		}
	}

	b.log.Info("initializing services...")

	// 1. Logger: MUST be first (synchronous)
	if b.logger == nil {
		if b.loggerConf == nil {
			b.loggerConf = &logger.Config{
				IsLocal: true,
			}
		}
		l, err := logger.New(b.loggerConf)
		if err != nil {
			return err
		}
		b.logger = l
	}

	// 2. Setup server config EARLY (before parallel init)
	if b.serverConf == nil {
		b.serverConf = &server.Config{}
	}

	if b.serverConf.Shutdown == nil {
		b.serverConf.Shutdown = func(ctx context.Context) error {
			b.closeServices()
			return nil
		}
	}

	// run services
	if err := b.runServices(ctx); err != nil {
		return nil
	}

	// Consume after init hooks (services are now available, can set Setup/Shutdown here)
	for _, fn := range b.afterInit {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("[bootstrap] after init hook failed: %w", err)
		}
	}

	b.mu.Lock()
	b.initialized = true
	b.mu.Unlock()

	b.log.Info("initialization completed")
	return nil
}

func (b *Bootstrap) runServices(ctx context.Context) error {
	// Init services in parallel
	g, ctx := errgroup.WithContext(ctx)

	// DB
	if b.dbConf != nil && b.database == nil {
		g.Go(func() error {
			db, err := database.New(b.dbConf)
			if err != nil {
				return err
			}
			b.database = db
			return nil
		})
	}

	// Redis
	if b.redisConf != nil && b.redisCache == nil {
		g.Go(func() error {
			cache, err := redis.New(b.redisConf)
			if err != nil {
				return fmt.Errorf("[redis] %w", err)
			}
			b.redisCache = cache
			return nil
		})
	}

	// MQ
	if b.rabbitmqConf != nil && b.rabbitmq == nil {
		g.Go(func() error {
			mq, err := rabbitmq.New(b.rabbitmqConf)
			if err != nil {
				return fmt.Errorf("[rabbitmq] %w", err)
			}
			b.rabbitmq = mq
			return nil
		})
	}

	// Keycloak
	if b.keycloakConf != nil && b.keycloak == nil {
		g.Go(func() error {
			b.keycloak = keycloak.New(b.keycloakConf)
			return nil
		})
	}

	// Scheduler
	if b.schedulerOn && b.scheduler == nil {
		g.Go(func() error {
			b.scheduler = scheduler.New(b.schedulerOpt...)
			return nil
		})
	}

	// REST client: init after logger is ready (may need logger)
	// If logger is not in options, inject it automatically
	if b.restOn && b.restClient == nil {
		g.Go(func() error {
			opts := b.restOpts
			// Check if WithLogger is already in options by checking if logger was passed
			// (rest.New will handle duplicates gracefully or user can avoid passing nil)
			if b.logger != nil {
				// Inject logger - if user passed nil, this will override it
				opts = append(opts, rest.WithLogger(b.logger))
			}
			b.restClient = rest.New(opts...)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

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

	// Consume before start hooks
	for _, fn := range b.beforeStart {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("[bootstrap] before start hook failed: %w", err)
		}
	}

	b.log.Info("starting services...")

	// Start scheduler if configured
	if b.scheduler != nil {
		b.scheduler.Start(ctx)
	}

	// Start HTTP server if configured
	if b.serverConf != nil {
		b.httpApp = server.New(b.serverConf)
		if err := b.httpApp.Start(); err != nil {
			return fmt.Errorf("[bootstrap] failed to start HTTP server: %w", err)
		}
		// Server errors are handled internally by HTTPApp
		// We don't need to monitor errCh separately since Start() is non-blocking
	}

	// Consume after start hooks
	for _, fn := range b.afterStart {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("after start hook failed: %w", err)
		}
	}

	b.mu.Lock()
	b.started = true
	b.mu.Unlock()

	b.log.Info("all services started")

	// Setup signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		b.log.Info("root context cancelled")
	case s := <-sig:
		b.log.Info("received signal: %v", s)
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

	b.log.Info("stopping services...")

	// Consume before stop hooks
	for _, fn := range b.beforeStop {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("[bootstrap] before stop hook error: %w", err)
		}
	}

	// Stop HTTP server if configured
	if b.httpApp != nil {
		if err := b.httpApp.Stop(ctx); err != nil {
			b.log.Info("HTTP server stop error: %v", err)
		}
	}

	// Close services
	b.closeServices()

	// Consume after stop hooks
	for _, fn := range b.afterStop {
		if err := fn(ctx); err != nil {
			b.log.Info("after stop hook error: %v", err)
		}
	}

	b.mu.Lock()
	b.started = false
	b.mu.Unlock()

	b.log.Info("all services stopped")
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

// Health checks the health of all configured services plus any custom health checkers
// registered via WithHealthChecker.
func (b *Bootstrap) Health(ctx context.Context) map[string]interface{} {
	health := make(map[string]interface{})

	if b.database != nil {
		if err := b.database.Ping(); err != nil {
			health["database"] = err
		} else {
			health["database"] = "OK"
		}
	}

	if b.redisCache != nil {
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := b.redisCache.Ping(ctxTimeout); err != nil {
			health["redis"] = err
		} else {
			health["redis"] = "OK"
		}
	}

	if b.rabbitmq != nil {
		conn, err := b.rabbitmq.GetConnection()
		if err != nil || conn == nil || conn.IsClosed() {
			health["rabbitmq"] = fmt.Errorf("connection not available")
		} else {
			health["rabbitmq"] = "OK"
		}
	}

	for _, entry := range b.healthCheckers {
		if err := entry.fn(ctx); err != nil {
			health[entry.name] = err
		} else {
			health[entry.name] = "OK"
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

func (b *Bootstrap) closeServices() {
	// Close Logger
	if b.logger != nil {
		b.logger.Sync()
		b.logger = nil
	}

	// Close DB
	if b.database != nil {
		b.database.Close()
		b.database = nil
	}

	// Close Redis
	if b.redisCache != nil {
		b.redisCache.Close()
		b.redisCache = nil
	}

	// Close MQ
	if b.rabbitmq != nil {
		b.rabbitmq.Close()
		b.rabbitmq = nil
	}

	// Close Kafka
	//if b.Kafka == nil {
	//	b.Kafka.Close()
	//	b.Kafka = nil
	//	b.KafkaProducer = nil
	//	b.KafkaConsumer = nil
	//} else if b.KafkaProducer != nil {
	//	b.KafkaProducer.Close()
	//	b.KafkaProducer = nil
	//} else if b.KafkaConsumer != nil {
	//	b.KafkaConsumer.Close()
	//	b.KafkaConsumer = nil
	//}
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

func (b *Bootstrap) RedisCache() *redis.Cache {
	return b.redisCache
}

func (b *Bootstrap) RESTClient() *rest.Client {
	return b.restClient
}

func (b *Bootstrap) Database() *database.DB {
	return b.database
}

func (b *Bootstrap) RabbitMQ() *rabbitmq.MQ {
	return b.rabbitmq
}

func (b *Bootstrap) KeyCloak() *keycloak.KC {
	return b.keycloak
}

func (b *Bootstrap) Logger() *logger.Logger {
	return b.logger
}

func (b *Bootstrap) Scheduler() *scheduler.Scheduler {
	return b.scheduler
}

func (b *Bootstrap) Migration() *migration.Migration {
	return b.migration
}

func (b *Bootstrap) Kafka() *kafkax.Kafka {
	return b.kafka
}
