package migration

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/BevisDev/godev/helper"
	"github.com/pressly/goose/v3"
)

type Migration struct {
	dir     string
	typeSQL string
	db      *sql.DB
}

func NewMigration(dir, typeSQL string, db *sql.DB) (*Migration, error) {
	m := Migration{
		dir:     dir,
		typeSQL: typeSQL,
		db:      db,
	}

	if err := m.Init(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Migration) Init() error {
	var dialect string
	switch m.typeSQL {
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
	if _, err := os.Stat(m.dir); os.IsNotExist(err) {
		return err
	}

	goose.SetTableName("db_version")
	return nil
}

func (m *Migration) Up(ctx context.Context, version int64) error {
	if version != 0 {
		if err := goose.UpToContext(ctx, m.db, m.dir, version); err != nil {
			return err
		}
	} else {
		if err := goose.UpContext(ctx, m.db, m.dir); err != nil {
			return err
		}
	}

	if err := goose.Status(m.db, m.dir); err != nil {
		return err
	}
	return nil
}

func (m *Migration) Down(ctx context.Context, version int64) error {
	if version != 0 {
		if err := goose.DownToContext(ctx, m.db, m.dir, version); err != nil {
			return err
		}
	} else {
		if err := goose.DownContext(ctx, m.db, m.dir); err != nil {
			return err
		}
	}

	if err := goose.Status(m.db, m.dir); err != nil {
		return err
	}
	return nil
}
