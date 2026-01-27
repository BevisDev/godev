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

	"github.com/BevisDev/godev/kafkax"

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

	// Core services
	Logger        *logger.Logger
	Database      *database.Database
	Migration     migration.Migration
	Redis         *redis.Cache
	RabbitMQ      *rabbitmq.RabbitMQ
	Keycloak      *keycloak.KeyCloak
	Kafka         *kafkax.Kafka
	KafkaProducer *kafkax.Producer
	KafkaConsumer *kafkax.Consumer
	Rest          *rest.Client
	Scheduler     *scheduler.Scheduler
	HTTPApp       *server.HTTPApp

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

	// Consume before init hooks
	for _, fn := range b.beforeInit {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("[bootstrap] before init failed: %w", err)
		}
	}

	log.Println("[bootstrap] initializing services...")

	// Logger: must be initialized first (synchronously) for other services
	if b.Logger == nil {
		if b.loggerConf == nil {
			b.loggerConf = &logger.Config{
				IsLocal: true,
			}
		}
		l, err := logger.New(b.loggerConf)
		if err != nil {
			return err
		}
		b.Logger = l
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

	// Kafka
	//if b.kafkaConf != nil && b.Kafka == nil {
	//	g.Go(func() error {
	//		kafka, err := kafkax.New(b.kafkaConf)
	//		if err != nil {
	//			return fmt.Errorf("[kafka] %w", err)
	//		}
	//		b.Kafka = kafka
	//		return nil
	//	})
	//} else if b.kafkaProducerConf != nil && b.KafkaProducer == nil {
	//	g.Go(func() error {
	//		p, err := kafkax.NewProducer(b.kafkaProducerConf)
	//		if err != nil {
	//			return fmt.Errorf("[kafka-producer] %w", err)
	//		}
	//		b.KafkaProducer = p
	//		return nil
	//	})
	//} else if b.kafkaConsumerConf != nil && b.KafkaConsumer == nil {
	//	g.Go(func() error {
	//		c, err := kafkax.NewConsumer(b.kafkaConsumerConf)
	//		if err != nil {
	//			return fmt.Errorf("[kafka-consumer] %w", err)
	//		}
	//		b.KafkaConsumer = c
	//		return nil
	//	})
	//}

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

	// REST client: init after logger is ready (may need logger)
	// If logger is not in options, inject it automatically
	if b.restOn && b.Rest == nil {
		g.Go(func() error {
			opts := b.restOpts
			// Check if WithLogger is already in options by checking if logger was passed
			// (rest.New will handle duplicates gracefully or user can avoid passing nil)
			if b.Logger != nil {
				// Inject logger - if user passed nil, this will override it
				opts = append(opts, rest.WithLogger(b.Logger))
			}
			b.Rest = rest.New(opts...)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Consume after init hooks (services are now available, can set Setup/Shutdown here)
	for _, fn := range b.afterInit {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("[bootstrap] after init hook failed: %w", err)
		}
	}

	// Ensure server config exists (for setting Setup/Shutdown later)
	if b.serverConf == nil {
		b.serverConf = &server.Config{
			Shutdown: func(ctx context.Context) error {
				b.closeServices()
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

	// Consume before start hooks
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

	// Consume after start hooks
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

	// Consume before stop hooks
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

	// Close services
	b.closeServices()

	// Consume after stop hooks
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

// Health checks the health of all configured services plus any custom health checkers
// registered via WithHealthChecker.
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
	if b.Logger != nil {
		b.Logger.Sync()
		b.Logger = nil
	}

	// Close Database
	if b.Database != nil {
		b.Database.Close()
		b.Database = nil
	}

	// Close Redis
	if b.Redis != nil {
		b.Redis.Close()
		b.Redis = nil
	}

	// Close RabbitMQ
	if b.RabbitMQ != nil {
		b.RabbitMQ.Close()
		b.RabbitMQ = nil
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
