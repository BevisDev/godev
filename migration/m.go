package migration

import (
	"context"
	"errors"
	"os"

	"github.com/BevisDev/godev/utils"
	"github.com/pressly/goose/v3"
)

// migration handles the setup and execution of database migrations using the Goose migration tool.
//
// It holds configuration for the migration directory, target database type, and the active *sql.DB connection.
// The migration dialect and working directory are initialized via the Init method.
type migration struct {
	*Config
}

// New creates a new Migration instance with the given migration directory,
// database kind, and sql.DB connection.
//
// It initializes the migration environment by setting the dialect and checking the
// existence of the migration folder.
//
// Returns an error if initialization fails (e.g., missing directory or invalid dialect).
func New(cf *Config) (Migration, error) {
	if cf == nil {
		return nil, errors.New("config is nil")
	}
	cf.withDefaults()

	m := migration{Config: cf}

	if err := m.Init(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *migration) Init() error {
	dialect := m.DBType.GetDialect()
	if dialect == "" {
		return errors.New("db type is unsupported")
	}

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}
	if _, err := os.Stat(m.Dir); os.IsNotExist(err) {
		return err
	}
	goose.SetTableName("db_version")
	return nil
}

func (m *migration) Status() error {
	return goose.Status(m.DB, m.Dir)
}

func (m *migration) Up(c context.Context, version int64) error {
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

func (m *migration) Down(c context.Context, version int64) error {
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
