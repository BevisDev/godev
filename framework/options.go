package framework

import (
	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/ginfw/server"
	"github.com/BevisDev/godev/keycloak"
	"github.com/BevisDev/godev/logx"
	"github.com/BevisDev/godev/rabbitmq"
	"github.com/BevisDev/godev/redis"
	"github.com/BevisDev/godev/rest"
	"github.com/BevisDev/godev/scheduler"
)

// Option configures Bootstrap behavior (captures config to initialize later in Init).
type Option func(*Bootstrap)

// WithLogger configures the logger.
func WithLogger(cfg *logx.Config) Option {
	return func(b *Bootstrap) {
		b.loggerCfg = cfg
	}
}

// WithDatabase configures the database connection.
func WithDatabase(cfg *database.Config) Option {
	return func(b *Bootstrap) {
		b.dbCfg = cfg
	}
}

// WithRedis configures the Redis cache.
func WithRedis(cfg *redis.Config) Option {
	return func(b *Bootstrap) {
		b.redisCfg = cfg
	}
}

// WithRabbitMQ configures RabbitMQ connection.
func WithRabbitMQ(cfg *rabbitmq.Config) Option {
	return func(b *Bootstrap) {
		b.rabbitmqCfg = cfg
	}
}

// WithKeycloak configures Keycloak client.
func WithKeycloak(cfg *keycloak.Config) Option {
	return func(b *Bootstrap) {
		b.keycloakCfg = cfg
	}
}

// WithRestClient configures REST HTTP client.
func WithRestClient(opts ...rest.OptionFunc) Option {
	return func(b *Bootstrap) {
		b.restInit = true
		b.restOpts = opts
	}
}

// WithScheduler configures the job scheduler.
func WithScheduler(opts ...scheduler.OptionFunc) Option {
	return func(b *Bootstrap) {
		b.schedulerInit = true
		b.schedulerOpt = opts
	}
}

// WithServer configures the Gin HTTP server.
func WithServer(cfg *server.Config) Option {
	return func(b *Bootstrap) {
		b.serverCfg = cfg
	}
}
