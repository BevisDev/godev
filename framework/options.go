package framework

import (
	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/ginfw/server"
	"github.com/BevisDev/godev/keycloak"
	"github.com/BevisDev/godev/logx"
	"github.com/BevisDev/godev/migration"
	"github.com/BevisDev/godev/rabbitmq"
	"github.com/BevisDev/godev/redis"
	"github.com/BevisDev/godev/rest"
	"github.com/BevisDev/godev/scheduler"
)

// Option configures Bootstrap behavior (captures config to initialize later in Init).
type Option func(*options)

type options struct {
	loggerConf    *logx.Config
	dbConf        *database.Config
	migrationConf *migration.Config
	redisConf     *redis.Config
	rabbitmqConf  *rabbitmq.Config
	keycloakConf  *keycloak.Config

	restOn   bool
	restOpts []rest.Option

	schedulerOn  bool
	schedulerOpt []scheduler.Option

	serverConf *server.Config
}

// WithLogger configures the logger.
func WithLogger(cfg *logx.Config) Option {
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
