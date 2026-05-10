package providers

import (
	"context"

	"github.com/BevisDev/godev/migration"
)

type MigrationProvider struct {
	cfg *migration.Config
	m   *migration.Migration
}

func NewMigrationProvider(cfg *migration.Config) *MigrationProvider {
	return &MigrationProvider{cfg: cfg}
}

func (p *MigrationProvider) Init(ctx context.Context) error {
	_ = ctx
	m, err := migration.New(p.cfg)
	if err != nil {
		return err
	}
	p.m = m
	return nil
}

func (p *MigrationProvider) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *MigrationProvider) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *MigrationProvider) Migration() *migration.Migration {
	return p.m
}
