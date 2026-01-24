package framework

import (
	"context"

	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/ginfw/server"
	"github.com/BevisDev/godev/kafkax"
	"github.com/BevisDev/godev/keycloak"
	"github.com/BevisDev/godev/logger"
	"github.com/BevisDev/godev/migration"
	"github.com/BevisDev/godev/rabbitmq"
	"github.com/BevisDev/godev/redis"
	"github.com/BevisDev/godev/rest"
	"github.com/BevisDev/godev/scheduler"
)

// Option configures Bootstrap behavior (captures config to initialize later in Init).
type Option func(*options)

// HealthChecker is a function that checks health of a service. Return nil if OK, otherwise an error.
type HealthChecker func(ctx context.Context) error

type healthCheckerEntry struct {
	name string
	fn   HealthChecker
}

type options struct {
	loggerConf    *logger.Config
	dbConf        *database.Config
	migrationConf *migration.Config
	redisConf     *redis.Config
	rabbitmqConf  *rabbitmq.Config
	keycloakConf  *keycloak.Config

	// kafka
	kafkaConf         *kafkax.Config
	kafkaProducerConf *kafkax.Config
	kafkaConsumerConf *kafkax.Config

	restOn   bool
	restOpts []rest.Option

	schedulerOn  bool
	schedulerOpt []scheduler.Option

	serverConf *server.Config

	// custom health checkers (e.g. from other projects)
	healthCheckers []healthCheckerEntry
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
func WithRabbitMQ(cfg *rabbitmq.Config) Option {
	return func(o *options) {
		o.rabbitmqConf = cfg
	}
}

// WithKeycloak configures Keycloak client.
func WithKeycloak(cfg *keycloak.Config) Option {
	return func(o *options) {
		o.keycloakConf = cfg
	}
}

// WithRestClient configures REST HTTP client.
func WithRestClient(opts ...rest.Option) Option {
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

// WithKafka configures the Kafka connection.
func WithKafka(cfg *kafkax.Config) Option {
	return func(o *options) {
		o.kafkaConf = cfg
	}
}

// WithKafkaProducer configures the Kafka Producer connection.
func WithKafkaProducer(cfg *kafkax.Config) Option {
	return func(o *options) {
		o.kafkaProducerConf = cfg
	}
}

// WithKafkaConsumer configures the Kafka Consumer connection.
func WithKafkaConsumer(cfg *kafkax.Config) Option {
	return func(o *options) {
		o.kafkaConsumerConf = cfg
	}
}

// WithHealthChecker registers a custom health checker. Name is used as the key in Health() result.
// Use this to plug in health checks from other projects (e.g. external APIs, custom services).
func WithHealthChecker(name string, fn HealthChecker) Option {
	return func(o *options) {
		if name != "" && fn != nil {
			o.healthCheckers = append(o.healthCheckers, healthCheckerEntry{name: name, fn: fn})
		}
	}
}
