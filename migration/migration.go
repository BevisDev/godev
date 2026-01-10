package migration

import (
	"context"
	"database/sql"
	"os"

	"github.com/BevisDev/godev/utils"
	"github.com/pressly/goose/v3"
)

// Config holds database migration settings.
//
// Dir specifies the directory containing migration scripts.
// Kind defines the type of database (e.g., Postgres, MySQL, SQLServer).
// DB is the active database connection used for applying migrations.
// Timeout sets the maximum duration (in seconds) allowed for each migration operation.
type Config struct {
	Dir     string
	DBType  DBType
	DB      *sql.DB
	Timeout int
}

// Migration handles the setup and execution of database migrations using the Goose migration tool.
//
// It holds configuration for the migration directory, target database type, and the active *sql.DB connection.
// The migration dialect and working directory are initialized via the Init method.
type Migration struct {
	*Config
}

const (
	// defaultTimeoutSec defines the default timeout (in seconds) for migration operations.
	defaultTimeout = 60
)

// New creates a new Migration instance with the given migration directory,
// database kind, and sql.DB connection.
//
// It initializes the migration environment by setting the dialect and checking the
// existence of the migration folder.
//
// Returns an error if initialization fails (e.g., missing directory or invalid dialect).
func New(cf *Config) (Exec, error) {
	// set default timeout
	if cf.Timeout <= 0 {
		cf.Timeout = defaultTimeout
	}

	m := Migration{Config: cf}

	if err := m.Init(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *Migration) Init() error {
	if err := goose.SetDialect(m.DBType.GetDialect()); err != nil {
		return err
	}
	if _, err := os.Stat(m.Dir); os.IsNotExist(err) {
		return err
	}
	goose.SetTableName("db_version")
	return nil
}

func (m *Migration) Status() error {
	return goose.Status(m.DB, m.Dir)
}

func (m *Migration) Up(c context.Context, version int64) error {
	ctx, cancel := utils.NewCtxTimeout(c, m.Timeout)
	defer cancel()

	var err error
	if version != 0 {
		err = goose.UpToContext(ctx, m.DB, m.Dir, version)
	} else {
		err = goose.UpContext(ctx, m.DB, m.Dir)
	}

	if err != nil {
		return err
	}

	return m.Status()
}

func (m *Migration) Down(c context.Context, version int64) error {
	ctx, cancel := utils.NewCtxTimeout(c, m.Timeout)
	defer cancel()

	var err error
	if version != 0 {
		err = goose.DownToContext(ctx, m.DB, m.Dir, version)
	} else {
		err = goose.DownContext(ctx, m.DB, m.Dir)
	}

	if err != nil {
		return err
	}

	return m.Status()
}
