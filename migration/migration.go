package migration

import (
	"context"
	"database/sql"
	"github.com/BevisDev/godev/utils"
	"github.com/pressly/goose/v3"
	"os"
)

// Migration handles the setup and execution of database migrations using the Goose migration tool.
//
// It holds configuration for the migration directory, target database type, and the active *sql.DB connection.
// The migration dialect and working directory are initialized via the Init method.
type Migration struct {
	dir     string
	kind    KindDB
	db      *sql.DB
	timeout int
	config  *Config
}

const (
	// defaultTimeoutSec defines the default timeout (in seconds) for migration operations.
	defaultTimeout = 60
)

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
func NewMigration(cf *Config) (*Migration, error) {
	// set default timeout
	if cf.Timeout <= 0 {
		cf.Timeout = defaultTimeout
	}

	m := Migration{
		dir:     cf.Dir,
		kind:    cf.Kind,
		db:      cf.DB,
		timeout: cf.Timeout,
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

func (m *Migration) Status() error {
	return goose.Status(m.db, m.dir)
}

func (m *Migration) SetTimeout(timeoutSec int) {
	m.timeout = timeoutSec
}

func (m *Migration) Up(c context.Context, version int64) error {
	ctx, cancel := utils.NewCtxTimeout(c, m.timeout)
	defer cancel()

	var err error
	if version != 0 {
		err = goose.UpToContext(ctx, m.db, m.dir, version)
	} else {
		err = goose.UpContext(ctx, m.db, m.dir)
	}

	if err != nil {
		return err
	}

	return m.Status()
}

func (m *Migration) Down(c context.Context, version int64) error {
	ctx, cancel := utils.NewCtxTimeout(c, m.timeout)
	defer cancel()

	var err error
	if version != 0 {
		err = goose.DownToContext(ctx, m.db, m.dir, version)
	} else {
		err = goose.DownContext(ctx, m.db, m.dir)
	}

	if err != nil {
		return err
	}

	return m.Status()
}
