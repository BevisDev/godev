package providers

import (
	"context"

	"github.com/BevisDev/godev/mailer"
)

type MailerProvider struct {
	cfg *mailer.Config
	m   mailer.Mailer
}

func NewMailerProvider(cfg *mailer.Config) *MailerProvider {
	return &MailerProvider{cfg: cfg}
}

func (p *MailerProvider) Init(ctx context.Context) error {
	_ = ctx
	m, err := mailer.New(p.cfg)
	if err != nil {
		return err
	}
	p.m = m
	return nil
}

func (p *MailerProvider) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *MailerProvider) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *MailerProvider) Mailer() mailer.Mailer {
	return p.m
}
