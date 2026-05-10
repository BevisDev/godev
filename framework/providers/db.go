package providers

import (
	"context"

	"github.com/BevisDev/godev/database"
)

type DBProvider struct {
	cfg *database.Config
	db  *database.DB
}

func NewDBProvider(cfg *database.Config) *DBProvider {
	return &DBProvider{cfg: cfg}
}

func (p *DBProvider) Init(ctx context.Context) error {
	_ = ctx
	db, err := database.New(p.cfg)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

func (p *DBProvider) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *DBProvider) Stop(ctx context.Context) error {
	_ = ctx
	if p.db != nil {
		p.db.Close()
	}
	return nil
}

func (p *DBProvider) DB() *database.DB {
	return p.db
}
