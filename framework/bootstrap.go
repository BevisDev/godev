package framework

import (
	"context"
	"fmt"
	"sync"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/console"

	"github.com/BevisDev/godev/ginfw/server"
	"github.com/gin-gonic/gin"
)

const bootstrap = "bootstrap"
const prefixBootstrap = "[" + bootstrap + "] "

type Hook func(ctx context.Context) error

// Bootstrap manages application lifecycle and dependencies.
type Bootstrap struct {
	log *console.Logger

	// service
	svcConf *ServiceConf
	svc     *Service

	// Lifecycle
	init    *Initializer
	starter *Starter
	stopper *Stopper

	// server
	serverConf *server.Config
	httpApp    *server.HTTPApp

	// Internal state
	mu      sync.RWMutex
	started bool

	ctx    context.Context
	cancel context.CancelFunc

	healthCheckers []healthChecker
}

// New creates a new Bootstrap instance with the provided options.
func New(c context.Context, opts ...Option) *Bootstrap {
	ctx, cancel := utils.NewCtxCancel(c)
	b := &Bootstrap{
		log:     console.New(bootstrap),
		ctx:     ctx,
		cancel:  cancel,
		svcConf: &ServiceConf{},
	}
	b.init = NewInitializer(b)
	b.starter = NewStarter(b)
	b.stopper = NewStopper(b)

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Run initializes, starts, and manages the application lifecycle.
// It blocks until a shutdown signal is received, then gracefully stops all services.
func (b *Bootstrap) Run(ctx context.Context) error {
	if err := b.init.Run(ctx); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Start and wait
	if err := b.starter.Start(ctx); err != nil {
		_ = b.Stop(ctx)
		return fmt.Errorf("start failed: %w", err)
	}

	return b.Stop(ctx)
}

// Context returns the bootstrap context.
func (b *Bootstrap) Context() context.Context {
	return b.ctx
}

// Shutdown triggers graceful shutdown.
func (b *Bootstrap) Shutdown() {
	b.cancel()
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

// Init returns the initialization phase for registering hooks.
// Example: b.Init().Before(setupLogger)
func (b *Bootstrap) Init() *Initializer { return b.init }

// Starter returns the start phase for registering hooks.
func (b *Bootstrap) Starter() *Starter { return b.starter }

// Stopper returns the stop phase for registering hooks.
func (b *Bootstrap) Stopper() *Stopper { return b.stopper }

// === Convenience hook shortcuts ===

func (b *Bootstrap) BeforeInit(fn Hook) error  { return b.init.Before(fn) }
func (b *Bootstrap) AfterInit(fn Hook) error   { return b.init.After(fn) }
func (b *Bootstrap) BeforeStart(fn Hook) error { return b.starter.Before(fn) }
func (b *Bootstrap) AfterStart(fn Hook) error  { return b.starter.After(fn) }
func (b *Bootstrap) BeforeStop(fn Hook) error  { return b.stopper.Before(fn) }
func (b *Bootstrap) AfterStop(fn Hook) error   { return b.stopper.After(fn) }

// RegisterProviders binds providers to lifecycle hooks:
// - BeforeInit:  provider.Init
// - BeforeStart: provider.Start
// - BeforeStop:  provider.Stop (reverse order)
func (b *Bootstrap) RegisterProviders(providers ...Provider) error {
	valid := make([]Provider, 0, len(providers))
	for _, p := range providers {
		if p != nil {
			valid = append(valid, p)
		}
	}
	if len(valid) == 0 {
		return nil
	}

	if err := b.BeforeInit(func(ctx context.Context) error {
		for idx, p := range valid {
			if err := p.Init(ctx); err != nil {
				return fmt.Errorf("%sprovider init[%d] (%T): %w", prefixBootstrap, idx, p, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := b.BeforeStart(func(ctx context.Context) error {
		for idx, p := range valid {
			if err := p.Start(ctx); err != nil {
				return fmt.Errorf("%sprovider start[%d] (%T): %w", prefixBootstrap, idx, p, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := b.BeforeStop(func(ctx context.Context) error {
		for idx := len(valid) - 1; idx >= 0; idx-- {
			p := valid[idx]
			if err := p.Stop(ctx); err != nil {
				return fmt.Errorf("%sprovider stop[%d] (%T): %w", prefixBootstrap, idx, p, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (b *Bootstrap) closeServices() {
	if b.svc != nil {
		b.svc.CloseAll()
	}
}
