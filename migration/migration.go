package migration

import (
	"context"
	"database/sql"
	"os"

	"github.com/pressly/goose/v3"
)

type Migration struct {
	dir  string
	kind KindDB
	db   *sql.DB
}

type KindDB string

const (
	SQLServer KindDB = "sqlserver"
	Postgres  KindDB = "postgres"
)

var dialectMap = map[KindDB]string{
	SQLServer: "mssql",
	Postgres:  "postgres",
}

func NewMigration(dir string, kind KindDB, db *sql.DB) (*Migration, error) {
	m := Migration{
		dir:  dir,
		kind: kind,
		db:   db,
	}

	if err := m.Init(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Migration) Init() error {
	if err := goose.SetDialect(dialectMap[m.kind]); err != nil {
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
