package database

// Config defines the configuration for connecting to a SQL database.
//
// It supports common settings such as connection parameters, pool sizing, timeouts,
// and connection string customization. The `DBType` field determines the target
// database type (e.g., MySQL, Postgres, SQL Server, Oracle).
type Config struct {
	// Kind specifies the type of database (e.g., types.MySQL, types.Postgres).
	DBType DBType

	// DBName is the name of the target database
	DBName string

	// TimeoutSec defines the default timeout (in seconds) for DB operations.
	TimeoutSec int

	// Host is the hostname or IP address of the database server.
	Host string

	// Port is the port number to connect to the database server.
	Port int

	// Username is the database login username.
	Username string

	// Password is the database login password.
	Password string

	// MaxOpenConns sets the maximum number of open connections to the database.
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections in the pool.
	MaxIdleConns int

	// MaxIdleTimeSec is the maximum amount of time (in seconds) a connection can remain idle.
	MaxIdleTimeSec int

	// MaxLifeTimeSec is the maximum amount of time (in seconds) a connection can be reused.
	MaxLifeTimeSec int

	// ShowQuery enables SQL query logging when set to true.
	ShowQuery bool

	// Params is an optional map of additional connection string parameters.
	Params map[string]string
}
