package migration

import (
	"context"
	"errors"
	"os"

	"github.com/BevisDev/godev/utils"
	"github.com/pressly/goose/v3"
)

// Migration handles the setup and execution of database migrations using the Goose Migration tool.
//
// It holds configuration for the migration directory, target database type, and the active *sql.DB connection.
// The Migration dialect and working directory are initialized via the Init method.
type Migration struct {
	cf *Config
}

// New creates a new Migration instance with the given Migration directory,
// database kind, and sql.DB connection.
//
// It initializes the Migration environment by setting the dialect and checking the
// existence of the Migration folder.
//
// Returns an error if initialization fails (e.g., missing directory or invalid dialect).
func New(cf *Config) (*Migration, error) {
	if cf == nil {
		return nil, errors.New("[migration] config is nil")
	}
	cf.withDefaults()

	m := Migration{cf: cf}

	if err := m.Init(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Init sets up the Migration system.
func (m *Migration) Init() error {
	dialect := m.cf.DBType.GetDialect()
	if dialect == "" {
		return errors.New("[migration] db type is not supported")
	}

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if _, err := os.Stat(m.cf.Dir); os.IsNotExist(err) {
		return err
	}
	goose.SetTableName("db_version")
	return nil
}

// Status retrieves the current Migration status.
func (m *Migration) Status() error {
	return goose.Status(m.cf.DB, m.cf.Dir)
}

// Up applies migrations up to the given version.
// If version is 0, it may apply all pending migrations.
// The context can be used to cancel or timeout the operation.
// Returns an error if any Migration fails.
func (m *Migration) Up(c context.Context, version int64) error {
	ctx, cancel := utils.NewCtxTimeout(c, m.cf.Timeout)
	defer cancel()

	var err error
	if version != 0 {
		err = goose.UpToContext(ctx, m.cf.DB, m.cf.Dir, version)
	} else {
		err = goose.UpContext(ctx, m.cf.DB, m.cf.Dir)
	}

	if err != nil {
		return err
	}

	return m.Status()
}

// Down rolls back migrations down to the given version.
// If version is 0, it may rollback all applied migrations.
// The context can be used to cancel or timeout the operation.
// Returns an error if rollback fails.
func (m *Migration) Down(c context.Context, version int64) error {
	ctx, cancel := utils.NewCtxTimeout(c, m.cf.Timeout)
	defer cancel()

	var err error
	if version != 0 {
		err = goose.DownToContext(ctx, m.cf.DB, m.cf.Dir, version)
	} else {
		err = goose.DownContext(ctx, m.cf.DB, m.cf.Dir)
	}

	if err != nil {
		return err
	}

	return m.Status()
}
