package framework

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/BevisDev/godev/database"
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
	"golang.org/x/sync/errgroup"
)

type ServiceConf struct {
	loggerConf *logger.Config

	dbConf        *database.Config
	migrationConf *migration.Config
	keycloakConf  *keycloak.Config
	redisConf     *redis.Config

	tgBotConf *tgbot.Config
	tgBotOpt  []tgbot.Option

	rabbitConf *rabbitmq.Config
	rabbitOpt  []rabbitmq.Option

	kafkaConf            *kafkax.Config
	kafkaConsumerHandler kafkax.Handler
	kafkaConsumerRetry   struct {
		enabled    bool
		maxRetries int
		retryDelay time.Duration
	}

	restOn   bool
	restOpts []rest.Option

	mailerConf *mailer.Config

	schedulerOn  bool
	schedulerOpt []scheduler.Option
}

func NewServiceConf() *ServiceConf {
	return &ServiceConf{}
}

type Service struct {
	cfg *ServiceConf

	logger     *logger.Logger
	database   *database.DB
	migration  *migration.Migration
	redisCache *redis.Cache
	mailer     *mailer.Mailer
	rabbitmq   *rabbitmq.MQ
	keycloak   *keycloak.KC
	kafka      *kafkax.Kafka
	tgBot      *tgbot.TgBot
	restClient *rest.Client
	scheduler  *scheduler.Scheduler

	others    []Hook
	mu        sync.RWMutex
	closeOnce sync.Once
}

func NewService(_ context.Context, cfg *ServiceConf) *Service {
	if cfg == nil {
		cfg = NewServiceConf()
	}
	return &Service{cfg: cfg}
}

func (s *Service) Register(fn Hook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.others = append(s.others, fn)
}

func (s *Service) CloseAll() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.restClient != nil {
			if hc := s.restClient.GetClient(); hc != nil {
				if tr, ok := hc.Transport.(*http.Transport); ok {
					tr.CloseIdleConnections()
				}
			}
			s.restClient = nil
		}

		s.mailer = nil
		s.tgBot = nil
		s.keycloak = nil
		s.scheduler = nil
		s.migration = nil

		if s.logger != nil {
			s.logger.Sync()
			s.logger = nil
		}
		if s.database != nil {
			s.database.Close()
			s.database = nil
		}
		if s.redisCache != nil {
			s.redisCache.Close()
			s.redisCache = nil
		}
		if s.rabbitmq != nil {
			s.rabbitmq.Close()
			s.rabbitmq = nil
		}
		if s.kafka != nil {
			s.kafka.Close()
			s.kafka = nil
		}
	})
}

func (s *Service) Run(ctx context.Context) error {
	var initMu sync.Mutex
	g, gCtx := errgroup.WithContext(ctx)

	if s.logger == nil {
		g.Go(func() error {
			cfg := s.cfg.loggerConf
			if cfg == nil {
				cfg = &logger.Config{IsLocal: true}
			}
			l, err := logger.New(cfg)
			if err != nil {
				return err
			}
			initMu.Lock()
			s.logger = l
			initMu.Unlock()
			return nil
		})
	}

	for _, f := range s.others {
		fn := f
		g.Go(func() error {
			return fn(gCtx)
		})
	}

	if s.cfg.dbConf != nil && s.database == nil {
		g.Go(func() error {
			db, err := database.New(s.cfg.dbConf)
			if err != nil {
				return err
			}
			initMu.Lock()
			s.database = db
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.migrationConf != nil && s.migration == nil {
		g.Go(func() error {
			cfg := *s.cfg.migrationConf
			if cfg.DB == nil && s.database != nil && s.database.GetDB() != nil {
				cfg.DB = s.database.GetDB().DB
			}
			m, err := migration.New(&cfg)
			if err != nil {
				return err
			}
			initMu.Lock()
			s.migration = m
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.redisConf != nil && s.redisCache == nil {
		g.Go(func() error {
			cache, err := redis.New(s.cfg.redisConf)
			if err != nil {
				return fmt.Errorf("[redis] %w", err)
			}
			initMu.Lock()
			s.redisCache = cache
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.rabbitConf != nil && s.rabbitmq == nil {
		g.Go(func() error {
			mq, err := rabbitmq.New(ctx, s.cfg.rabbitConf, s.cfg.rabbitOpt...)
			if err != nil {
				return fmt.Errorf("[rabbitmq] %w", err)
			}
			initMu.Lock()
			s.rabbitmq = mq
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.mailerConf != nil && s.mailer == nil {
		g.Go(func() error {
			m, err := mailer.New(s.cfg.mailerConf)
			if err != nil {
				return fmt.Errorf("[mailer] %w", err)
			}
			initMu.Lock()
			s.mailer = m
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.keycloakConf != nil && s.keycloak == nil {
		g.Go(func() error {
			initMu.Lock()
			s.keycloak = keycloak.New(s.cfg.keycloakConf)
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.schedulerOn && s.scheduler == nil {
		g.Go(func() error {
			initMu.Lock()
			s.scheduler = scheduler.New(s.cfg.schedulerOpt...)
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.restOn && s.restClient == nil {
		g.Go(func() error {
			opts := append([]rest.Option{}, s.cfg.restOpts...)
			initMu.Lock()
			if s.logger != nil {
				opts = append(opts, rest.WithLogger(s.logger))
			}
			s.restClient = rest.New(opts...)
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.kafkaConf != nil && s.kafka == nil {
		g.Go(func() error {
			k, err := kafkax.New(s.cfg.kafkaConf)
			if err != nil {
				return err
			}
			initMu.Lock()
			s.kafka = k
			initMu.Unlock()
			return nil
		})
	}

	if s.cfg.tgBotConf != nil && s.tgBot == nil {
		g.Go(func() error {
			bot, err := tgbot.New(s.cfg.tgBotConf, s.cfg.tgBotOpt...)
			if err != nil {
				return err
			}
			initMu.Lock()
			s.tgBot = bot
			initMu.Unlock()
			return nil
		})
	}

	return g.Wait()
}

func (s *Service) RedisCache() *redis.Cache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.redisCache
}

func (s *Service) RESTClient() *rest.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.restClient
}

func (s *Service) Database() *database.DB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.database
}

func (s *Service) RabbitMQ() *rabbitmq.MQ {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rabbitmq
}

func (s *Service) KeyCloak() *keycloak.KC {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.keycloak
}

func (s *Service) Logger() *logger.Logger {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logger
}

func (s *Service) Scheduler() *scheduler.Scheduler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scheduler
}

func (s *Service) Migration() *migration.Migration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.migration
}

func (s *Service) Kafka() *kafkax.Kafka {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.kafka
}

func (s *Service) Mailer() *mailer.Mailer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mailer
}

func (s *Service) TgBot() *tgbot.TgBot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tgBot
}
