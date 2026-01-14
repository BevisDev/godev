package database

import (
	"fmt"
	"net/url"
	"time"
)

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

	// Timeout defines the default timeout for DB operations.
	Timeout time.Duration

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
	MaxIdleTimeSec time.Duration

	// MaxLifeTimeSec is the maximum amount of time (in seconds) a connection can be reused.
	MaxLifeTimeSec time.Duration

	// ShowQuery enables SQL query logging when set to true.
	ShowQuery bool

	// Params is an optional map of additional connection string parameters.
	Params map[string]string
}

const (
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

func (c *Config) withDefaults() {
	if c.Timeout <= 0 {
		c.Timeout = 1 * time.Minute
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 50
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 50
	}
	if c.MaxIdleTimeSec <= 0 {
		c.MaxIdleTimeSec = 5 * time.Second
	}
	if c.MaxLifeTimeSec <= 0 {
		c.MaxLifeTimeSec = 3600 * time.Second
	}
}

func (c *Config) getDSN() string {
	var template = c.DBType.ConnectionString()
	if template == "" {
		return ""
	}

	var connStr string
	switch c.DBType {
	case SqlServer:
		connStr = fmt.Sprintf(template,
			c.Username, c.Password, c.Host, c.Port, c.DBName)
		if len(c.Params) > 0 {
			params := url.Values{}
			for k, v := range c.Params {
				params.Add(k, v)
			}
			connStr += "&" + params.Encode()
		}
	case Postgres:
		connStr = fmt.Sprintf(template,
			c.Username, c.Password, c.Host, c.Port, c.DBName)
		if len(c.Params) > 0 {
			params := url.Values{}
			for k, v := range c.Params {
				params.Add(k, v)
			}
			connStr += "&" + params.Encode()
		}
	case Oracle:
		connStr = fmt.Sprintf(template,
			c.Username, c.Password, c.Host, c.Port, c.DBName)
	case MySQL:
		connStr = fmt.Sprintf(template,
			c.Username, c.Password, c.Host, c.Port, c.DBName)
		if len(c.Params) > 0 {
			params := url.Values{}
			for k, v := range c.Params {
				params.Add(k, v)
			}
			connStr += "?" + params.Encode()
		}
	default:
		return connStr
	}
	return connStr
}
