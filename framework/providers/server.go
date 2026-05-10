package providers

import (
	"context"

	"github.com/BevisDev/godev/ginfw/server"
)

type ServerProvider struct {
	cfg *server.Config
	app *server.HTTPApp
}

func NewServerProvider(cfg *server.Config) *ServerProvider {
	return &ServerProvider{cfg: cfg}
}

func (p *ServerProvider) Init(ctx context.Context) error {
	_ = ctx
	p.app = server.New(p.cfg)
	return nil
}

func (p *ServerProvider) Start(ctx context.Context) error {
	_ = ctx
	if p.app == nil {
		return nil
	}
	return p.app.Start()
}

func (p *ServerProvider) Stop(ctx context.Context) error {
	if p.app == nil {
		return nil
	}
	return p.app.Stop(ctx)
}

func (p *ServerProvider) App() *server.HTTPApp {
	return p.app
}
