package redis

import (
	"fmt"
	"time"
)

// Config holds configuration options for connecting to a Redis instance.
//
// It includes host address, port, authentication credentials, selected DB index,
// connection pool size, and a default timeout for Redis operations.
type Config struct {
	Host     string        // Redis server hostname or IP
	Port     int           // Redis server port
	Password string        // Password for authentication (if required)
	DB       int           // Redis database index (0 by default)
	PoolSize int           // Maximum number of connections in the pool
	Timeout  time.Duration // timeout for Redis operations in seconds
}

func (c *Config) withDefaults() {
	if c.Timeout <= 0 {
		c.Timeout = 1 * time.Minute
	}
	if c.PoolSize <= 0 {
		c.PoolSize = 10
	}
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
