package framework

import (
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
type Option func(*Bootstrap)

type healthChecker struct {
	name string
	fn   Hook
}

type options struct {
	_ struct{}
}

// WithLogger configures the logger.
func WithLogger(cfg *logger.Config) Option {
	return func(o *Bootstrap) {
		o.svcConf.loggerConf = cfg
	}
}

// WithDatabase configures the database connection.
func WithDatabase(cfg *database.Config) Option {
	return func(b *Bootstrap) {
		b.svcConf.dbConf = cfg
	}
}

// WithMigration configures the database migration.
func WithMigration(cfg *migration.Config) Option {
	return func(b *Bootstrap) {
		b.svcConf.migrationConf = cfg
	}
}

// WithRedis configures the Redis cache.
func WithRedis(cfg *redis.Config) Option {
	return func(b *Bootstrap) {
		b.svcConf.redisConf = cfg
	}
}

// WithRabbitMQ configures RabbitMQ connection.
func WithRabbitMQ(cfg *rabbitmq.Config, opts ...rabbitmq.Option) Option {
	return func(b *Bootstrap) {
		b.svcConf.rabbitConf = cfg
		b.svcConf.rabbitOpt = append(b.svcConf.rabbitOpt, opts...)
	}
}

// WithMailer configures the mailer.
func WithMailer(cfg *mailer.Config) Option {
	return func(b *Bootstrap) {
		b.svcConf.mailerConf = cfg
	}
}

// WithTgBot configures the Telegram bot client to be initialized by Bootstrap.
func WithTgBot(cfg *tgbot.Config, opts ...tgbot.Option) Option {
	return func(b *Bootstrap) {
		b.svcConf.tgBotConf = cfg
		b.svcConf.tgBotOpt = append(b.svcConf.tgBotOpt, opts...)
	}
}

// WithKeycloak configures Keycloak client.
func WithKeycloak(cfg *keycloak.Config) Option {
	return func(b *Bootstrap) {
		b.svcConf.keycloakConf = cfg
	}
}

// WithRESTClient configures REST HTTP client.
func WithRESTClient(opts ...rest.Option) Option {
	return func(b *Bootstrap) {
		b.svcConf.restOn = true
		b.svcConf.restOpts = opts
	}
}

// WithScheduler configures the job scheduler.
func WithScheduler(opts ...scheduler.Option) Option {
	return func(b *Bootstrap) {
		b.svcConf.schedulerOn = true
		b.svcConf.schedulerOpt = opts
	}
}

// WithServer configures the Gin HTTP server.
func WithServer(cfg *server.Config) Option {
	return func(b *Bootstrap) {
		b.serverConf = cfg
	}
}

// WithKafka configures the Kafka connection.
func WithKafka(cfg *kafkax.Config) Option {
	return func(b *Bootstrap) {
		b.svcConf.kafkaConf = cfg
	}
}

// WithKafkaConsumer registers a handler to consume Kafka messages. The consumer loop is started
// automatically in Bootstrap.Start() when Kafka is configured with Consumer.GroupID and Consumer.Topics.
func WithKafkaConsumer(handler kafkax.Handler) Option {
	return func(b *Bootstrap) {
		b.svcConf.kafkaConsumerHandler = handler
	}
}

// WithKafkaConsumerRetry registers a handler with retry logic. The consumer loop is started
// automatically in Bootstrap.Start(). Failed messages are retried up to maxRetries with retryDelay between attempts.
func WithKafkaConsumerRetry(handler kafkax.Handler, maxRetries int, retryDelay time.Duration) Option {
	return func(b *Bootstrap) {
		b.svcConf.kafkaConsumerHandler = handler
		b.svcConf.kafkaConsumerRetry.enabled = true
		b.svcConf.kafkaConsumerRetry.maxRetries = maxRetries
		b.svcConf.kafkaConsumerRetry.retryDelay = retryDelay
	}
}

// WithHealthChecker registers a custom health checker. Name is used as the key in Health() result.
// Use this to plug in health checks from other projects (e.g. external APIs, custom services).
func WithHealthChecker(name string, fn Hook) Option {
	return func(b *Bootstrap) {
		if name != "" && fn != nil {
			b.healthCheckers = append(b.healthCheckers, healthChecker{name: name, fn: fn})
		}
	}
}
