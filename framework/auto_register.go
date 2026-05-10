package framework

import (
	"context"
	"time"

	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/framework/providers"
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
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type KafkaRetry struct {
	MaxRetries int
	RetryDelay time.Duration
}

// ProviderConfig lets external repos pass config only and auto-register providers.
type ProviderConfig struct {
	Logger *logger.Config

	Database  *database.Config
	Migration *migration.Config
	Redis     *redis.Config

	RabbitMQ        *rabbitmq.Config
	RabbitMQOptions []rabbitmq.Option

	Kafka        *kafkax.Config
	KafkaHandler kafkax.Handler
	KafkaRetry   *KafkaRetry

	Mailer   *mailer.Config
	Keycloak *keycloak.Config

	RESTOptions []rest.Option

	SchedulerOptions []scheduler.Option

	Server *server.Config

	TgBot        *tgbot.Config
	TgBotOptions []tgbot.Option
	TgBotHandler func(tgbotapi.Update)
}

// RegisterTo builds and registers providers based on non-nil configs.
func (c *ProviderConfig) RegisterTo(b *Bootstrap) error {
	if c == nil || b == nil {
		return nil
	}

	list := make([]Provider, 0, 12)

	if c.Logger != nil {
		list = append(list, providers.NewLoggerProvider(c.Logger))
	}
	if c.Database != nil {
		list = append(list, providers.NewDBProvider(c.Database))
	}
	if c.Migration != nil {
		list = append(list, providers.NewMigrationProvider(c.Migration))
	}
	if c.Redis != nil {
		list = append(list, providers.NewRedisProvider(c.Redis))
	}
	if c.RabbitMQ != nil {
		list = append(list, providers.NewRabbitMQProvider(c.RabbitMQ, c.RabbitMQOptions...))
	}
	if c.Kafka != nil {
		kp := providers.NewKafkaProvider(c.Kafka, c.KafkaHandler)
		if c.KafkaRetry != nil {
			kp.WithRetry(c.KafkaRetry.MaxRetries, c.KafkaRetry.RetryDelay)
		}
		list = append(list, kp)
	}
	if c.Mailer != nil {
		list = append(list, providers.NewMailerProvider(c.Mailer))
	}
	if c.Keycloak != nil {
		list = append(list, providers.NewKeycloakProvider(c.Keycloak))
	}
	if len(c.RESTOptions) > 0 {
		list = append(list, providers.NewRESTProvider(c.RESTOptions...))
	}
	if len(c.SchedulerOptions) > 0 {
		list = append(list, providers.NewSchedulerProvider(c.SchedulerOptions...))
	}
	if c.Server != nil {
		list = append(list, providers.NewServerProvider(c.Server))
	}
	if c.TgBot != nil {
		list = append(list, providers.NewTgBotProvider(c.TgBot, c.TgBotHandler, c.TgBotOptions...))
	}

	return b.RegisterProviders(list...)
}

// NewWithProviders creates bootstrap then auto-registers providers from config.
func NewWithProviders(ctx context.Context, cfg *ProviderConfig, opts ...Option) (*Bootstrap, error) {
	b := New(ctx, opts...)
	if err := cfg.RegisterTo(b); err != nil {
		return nil, err
	}
	return b, nil
}
