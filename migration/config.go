package migration

import (
	"database/sql"
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
