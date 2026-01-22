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

// OptionFunc configures Bootstrap behavior (captures config to initialize later in Init).
type OptionFunc func(*options)

type options struct {
	services int

	loggerConf    *logx.Config
	dbConf        *database.Config
	migrationConf *migration.Config
	redisConf     *redis.Config
	rabbitmqConf  *rabbitmq.Config
	keycloakConf  *keycloak.Config

	restOn   bool
	restOpts []rest.OptionFunc

	schedulerOn  bool
	schedulerOpt []scheduler.OptionFunc

	serverConf *server.Config
}

// WithLogger configures the logger.
func WithLogger(cfg *logx.Config) OptionFunc {
	return func(o *options) {
		o.loggerConf = cfg
	}
}

// WithDatabase configures the database connection.
func WithDatabase(cfg *database.Config) OptionFunc {
	return func(o *options) {
		o.dbConf = cfg
	}
}

// WithMigration configures the database migration.
func WithMigration(cfg *migration.Config) OptionFunc {
	return func(o *options) {
		o.migrationConf = cfg
	}
}

// WithRedis configures the Redis cache.
func WithRedis(cfg *redis.Config) OptionFunc {
	return func(o *options) {
		o.redisConf = cfg
	}
}

// WithRabbitMQ configures RabbitMQ connection.
func WithRabbitMQ(cfg *rabbitmq.Config) OptionFunc {
	return func(o *options) {
		o.rabbitmqConf = cfg
	}
}

// WithKeycloak configures Keycloak client.
func WithKeycloak(cfg *keycloak.Config) OptionFunc {
	return func(o *options) {
		o.keycloakConf = cfg
	}
}

// WithRestClient configures REST HTTP client.
func WithRestClient(opts ...rest.OptionFunc) OptionFunc {
	return func(o *options) {
		o.restOn = true
		o.restOpts = opts
	}
}

// WithScheduler configures the job scheduler.
func WithScheduler(opts ...scheduler.OptionFunc) OptionFunc {
	return func(o *options) {
		o.schedulerOn = true
		o.schedulerOpt = opts
	}
}

// WithServer configures the Gin HTTP server.
func WithServer(cfg *server.Config) OptionFunc {
	return func(o *options) {
		o.serverConf = cfg
	}
}
