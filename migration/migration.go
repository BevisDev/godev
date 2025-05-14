package migration

import (
	"context"
	"database/sql"
	"github.com/BevisDev/godev/types"
	"os"

	"github.com/pressly/goose/v3"
)

// Migration handles the setup and execution of database migrations using the Goose migration tool.
//
// It holds configuration for the migration directory, target database type, and the active *sql.DB connection.
// The migration dialect and working directory are initialized via the Init method.
type Migration struct {
	dir  string
	kind types.KindDB
	db   *sql.DB
}

// NewMigration creates a new Migration instance with the given migration directory,
// database kind, and sql.DB connection.
//
// It initializes the migration environment by setting the dialect and checking the
// existence of the migration folder.
//
// Returns an error if initialization fails (e.g., missing directory or invalid dialect).
//
// Example:
//
//	m, err := NewMigration("./migrations", types.Postgres, db)
//	if err != nil {
//	    log.Fatal("Failed to initialize migration:", err)
//	}
func NewMigration(dir string, kind types.KindDB, db *sql.DB) (*Migration, error) {
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
	if err := goose.SetDialect(m.kind.GetDialect()); err != nil {
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
