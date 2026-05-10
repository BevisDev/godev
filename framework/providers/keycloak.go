package providers

import (
	"context"

	"github.com/BevisDev/godev/keycloak"
)

type KeycloakProvider struct {
	cfg *keycloak.Config
	kc  *keycloak.KC
}

func NewKeycloakProvider(cfg *keycloak.Config) *KeycloakProvider {
	return &KeycloakProvider{cfg: cfg}
}

func (p *KeycloakProvider) Init(ctx context.Context) error {
	_ = ctx
	p.kc = keycloak.New(p.cfg)
	return nil
}

func (p *KeycloakProvider) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *KeycloakProvider) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *KeycloakProvider) KC() *keycloak.KC {
	return p.kc
}
