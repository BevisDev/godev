package migration

import (
	"context"
	"errors"
	"os"

	"github.com/BevisDev/godev/helper"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

type Migration struct {
	Dir        string
	Sqlx       *sqlx.DB
	TypeSQL    string
	TimeoutSec int
	Ctx        context.Context
	Version    int64
}

func (m *Migration) Init() error {
	var dialect string
	switch m.TypeSQL {
	case helper.SQLServer:
		dialect = "mssql"
		break
	case helper.Postgres:
		dialect = "postgres"
		break
	default:
		return errors.New("type SQL unsupported")
	}
	if err := goose.SetDialect(dialect); err != nil {
		return err
	}
	if _, err := os.Stat(m.Dir); os.IsNotExist(err) {
		return err
	}
	goose.SetTableName("db_version")
	if m.Ctx == nil {
		m.Ctx = context.Background()
	}
	return nil
}

func (m *Migration) Up() error {
	if err := m.Init(); err != nil {
		return err
	}
	ctx, cancel := helper.CreateCtxTimeout(m.Ctx, m.TimeoutSec)
	defer cancel()
	if m.Version != 0 {
		if err := goose.UpToContext(ctx, m.Sqlx.DB, m.Dir, m.Version); err != nil {
			return err
		}
	} else {
		if err := goose.UpContext(ctx, m.Sqlx.DB, m.Dir); err != nil {
			return err
		}
	}

	if err := goose.Status(m.Sqlx.DB, m.Dir); err != nil {
		return err
	}
	return nil
}

func (m *Migration) Down() error {
	if err := m.Init(); err != nil {
		return err
	}
	ctx, cancel := helper.CreateCtxTimeout(m.Ctx, m.TimeoutSec)
	defer cancel()
	if m.Version != 0 {
		if err := goose.DownToContext(ctx, m.Sqlx.DB, m.Dir, m.Version); err != nil {
			return err
		}
	} else {
		if err := goose.DownContext(ctx, m.Sqlx.DB, m.Dir); err != nil {
			return err
		}
	}

	if err := goose.Status(m.Sqlx.DB, m.Dir); err != nil {
		return err
	}
	return nil
}
