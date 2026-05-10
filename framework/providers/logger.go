package providers

import (
	"context"

	"github.com/BevisDev/godev/logger"
)

type LoggerProvider struct {
	cfg *logger.Config
	lg  *logger.Logger
}

func NewLoggerProvider(cfg *logger.Config) *LoggerProvider {
	return &LoggerProvider{cfg: cfg}
}

func (p *LoggerProvider) Init(ctx context.Context) error {
	_ = ctx
	lg, err := logger.New(p.cfg)
	if err != nil {
		return err
	}
	p.lg = lg
	return nil
}

func (p *LoggerProvider) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *LoggerProvider) Stop(ctx context.Context) error {
	_ = ctx
	if p.lg != nil {
		p.lg.Sync()
	}
	return nil
}

func (p *LoggerProvider) Logger() *logger.Logger {
	return p.lg
}
