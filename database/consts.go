package database

const (
	// DefaultTimeoutSec defines the default timeout (in seconds) for database operations.
	defaultTimeoutSec = 60

	// DefaultMaxOpenConn is the maximum number of open connections (default: 50).
	defaultMaxOpenConn = 50

	// DefaultMaxIdleConn is the maximum number of idle connections kept in the pool (default: 50).
	defaultMaxIdleConn = 50

	// DefaultConnMaxIdleTime is the max idle time in seconds before a connection is closed (default: 5s).
	defaultConnMaxIdleTime = 5

	// DefaultConnMaxLifetime is the max lifetime in seconds for a connection before recycling (default: 3600s / 1h).
	defaultConnMaxLifetime = 3600

	// MaxParams defines the maximum number of parameters allowed
	// To avoid hitting this hard limit, it's recommended to stay under 2000.
	// This value is used to determine safe batch sizes for bulk operations
	maxParams = 2000
)

var connectionMap = map[DBType]string{
	// username/password/host/port/db
	SqlServer: "sqlserver://%s:%s@%s:%d?database=%s",

	// username/password/host/port/db
	Postgres: "postgres://%s:%s@%s:%d/%s?sslmode=disable",

	// username/password@host:port/db
	Oracle: "%s/%s@%s:%d/%s",

	// username/password/host/port/db
	MySQL: "%s:%s@tcp(%s:%d)/%s",
}
