package providers

import (
	"context"
	"net/http"

	"github.com/BevisDev/godev/rest"
)

type RESTProvider struct {
	opts   []rest.Option
	client *rest.Client
}

func NewRESTProvider(opts ...rest.Option) *RESTProvider {
	return &RESTProvider{opts: opts}
}

func (p *RESTProvider) Init(_ context.Context) error {
	p.client = rest.New(p.opts...)
	return nil
}

func (p *RESTProvider) Start(_ context.Context) error {
	return nil
}

func (p *RESTProvider) Stop(ctx context.Context) error {
	_ = ctx
	if p.client != nil {
		if hc := p.client.GetClient(); hc != nil {
			if tr, ok := hc.Transport.(*http.Transport); ok {
				tr.CloseIdleConnections()
			}
		}
	}
	return nil
}

func (p *RESTProvider) Client() *rest.Client {
	return p.client
}
