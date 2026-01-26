package migration

import (
	"database/sql"
	"time"
)

// Config holds database Migration settings.
type Config struct {
	// Dir specifies the directory containing Migration scripts.
	Dir string

	// DBType defines the type of database (e.g., Postgres, MySQL, SQLServer).
	DBType DBType

	// DB is the active database connection used for applying migrations.
	DB *sql.DB

	// Timeout sets the maximum duration allowed for each Migration operation.
	Timeout time.Duration
}

func (c *Config) clone() *Config {
	cc := *c
	if cc.Timeout <= 0 {
		cc.Timeout = 1 * time.Minute
	}
	return &cc
}
