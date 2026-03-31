package framework

import (
	"context"
	"time"

	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/ginfw/server"
	"github.com/BevisDev/godev/kafkax"
	"github.com/BevisDev/godev/keycloak"
	"github.com/BevisDev/godev/logger"
	"github.com/BevisDev/godev/mailer"
	"github.com/BevisDev/godev/migration"
	"github.com/BevisDev/godev/rabbitmq"
	"github.com/BevisDev/godev/redis"
	"github.com/BevisDev/godev/rest"
	"github.com/BevisDev/godev/scheduler"
	"github.com/BevisDev/godev/tgbot"
)

// Option configures Bootstrap behavior (captures config to initialize later in Init).
type Option func(*options)

// HealthCheckFunc is a function that checks health of a service. Return nil if OK, otherwise an error.
type HealthCheckFunc func(ctx context.Context) error

type healthChecker struct {
	name string
	fn   HealthCheckFunc
}

type options struct {
	loggerConf *logger.Config

	// database
	dbConf        *database.Config
	migrationConf *migration.Config
	keycloakConf  *keycloak.Config
	redisConf     *redis.Config

	// tgbot
	tgBotConf *tgbot.Config
	tgBotOpt  []tgbot.Option

	// rabbitmq
	rabbitConf *rabbitmq.Config
	rabbitOpt  []rabbitmq.Option

	// kafka
	kafkaConf            *kafkax.Config
	kafkaConsumerHandler kafkax.Handler

	restOn   bool
	restOpts []rest.Option

	mailerConf *mailer.Config

	schedulerOn  bool
	schedulerOpt []scheduler.Option

	serverConf *server.Config

	// graceful shutdown for Run/Stop (also applied to server.Config.ShutdownTimeout when unset)
	shutdownTimeout time.Duration

	// custom health checkers (e.g. from other projects)
	healthCheckers []healthChecker
}

// WithLogger configures the logger.
func WithLogger(cfg *logger.Config) Option {
	return func(o *options) {
		o.loggerConf = cfg
	}
}

// WithDatabase configures the database connection.
func WithDatabase(cfg *database.Config) Option {
	return func(o *options) {
		o.dbConf = cfg
	}
}

// WithMigration configures the database migration.
func WithMigration(cfg *migration.Config) Option {
	return func(o *options) {
		o.migrationConf = cfg
	}
}

// WithRedis configures the Redis cache.
func WithRedis(cfg *redis.Config) Option {
	return func(o *options) {
		o.redisConf = cfg
	}
}

// WithRabbitMQ configures RabbitMQ connection.
func WithRabbitMQ(cfg *rabbitmq.Config, opts ...rabbitmq.Option) Option {
	return func(o *options) {
		o.rabbitConf = cfg
		o.rabbitOpt = append(o.rabbitOpt, opts...)
	}
}

// WithMailer configures the mailer.
func WithMailer(cfg *mailer.Config) Option {
	return func(o *options) {
		o.mailerConf = cfg
	}
}

// WithTgBot configures the Telegram bot client to be initialized by Bootstrap.
func WithTgBot(cfg *tgbot.Config, opts ...tgbot.Option) Option {
	return func(o *options) {
		o.tgBotConf = cfg
		o.tgBotOpt = append(o.tgBotOpt, opts...)
	}
}

// WithKeycloak configures Keycloak client.
func WithKeycloak(cfg *keycloak.Config) Option {
	return func(o *options) {
		o.keycloakConf = cfg
	}
}

// WithRESTClient configures REST HTTP client.
func WithRESTClient(opts ...rest.Option) Option {
	return func(o *options) {
		o.restOn = true
		o.restOpts = opts
	}
}

// WithScheduler configures the job scheduler.
func WithScheduler(opts ...scheduler.Option) Option {
	return func(o *options) {
		o.schedulerOn = true
		o.schedulerOpt = opts
	}
}

// WithServer configures the Gin HTTP server.
func WithServer(cfg *server.Config) Option {
	return func(o *options) {
		o.serverConf = cfg
	}
}

// WithShutdownTimeout sets the maximum duration for graceful shutdown (Bootstrap.Run / Stop,
// HTTP Shutdown hook, and http.Server.Shutdown). When > 0 it overrides server.Config.ShutdownTimeout
// during Init. Otherwise Bootstrap keeps your server timeout or applies a 15s default.
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = d
	}
}

// WithKafka configures the Kafka connection.
func WithKafka(cfg *kafkax.Config) Option {
	return func(o *options) {
		o.kafkaConf = cfg
	}
}

// WithKafkaConsumer registers a handler to consume Kafka messages. The consumer loop is started
// automatically in Bootstrap.Start() when Kafka is configured with Consumer.GroupID and Consumer.Topics.
// Retries are configured on cfg.Consumer (MaxHandlerRetries, HandlerRetryDelay), not on this option.
func WithKafkaConsumer(handler kafkax.Handler) Option {
	return func(o *options) {
		o.kafkaConsumerHandler = handler
	}
}

// WithHealthChecker registers a custom health checker. Name is used as the key in Health() result.
// Use this to plug in health checks from other projects (e.g. external APIs, custom services).
func WithHealthChecker(name string, fn HealthCheckFunc) Option {
	return func(o *options) {
		if name != "" && fn != nil {
			o.healthCheckers = append(o.healthCheckers, healthChecker{name: name, fn: fn})
		}
	}
}
