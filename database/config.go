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
	// DBType specifies the type of database (e.g., MySQL, Postgres, SqlServer, Oracle).
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

	// MaxIdleTime is the maximum amount of time a connection can remain idle.
	MaxIdleTime time.Duration

	// MaxLifeTime is the maximum amount of time a connection can be reused.
	MaxLifeTime time.Duration

	// ShowQuery enables SQL query logging when set to true.
	ShowQuery bool

	// Params is an optional map of additional connection string parameters.
	Params map[string]string
}

// clone applies default values to config fields if they are zero or invalid.
func (c *Config) clone() *Config {
	cc := *c
	if cc.Timeout <= 0 {
		cc.Timeout = 1 * time.Minute
	}
	if cc.MaxOpenConns <= 0 {
		cc.MaxOpenConns = 50
	}
	if cc.MaxIdleConns <= 0 {
		cc.MaxIdleConns = 50
	}
	if cc.MaxIdleTime <= 0 {
		cc.MaxIdleTime = 5 * time.Second
	}
	if cc.MaxLifeTime <= 0 {
		cc.MaxLifeTime = 3600 * time.Second
	}
	return &cc
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
