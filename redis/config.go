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

// clone applies default values to the configuration if they are not set.
func (c *Config) clone() *Config {
	clone := &Config{
		Host:     c.Host,
		Port:     c.Port,
		Password: c.Password,
		DB:       c.DB,
		PoolSize: c.PoolSize,
		Timeout:  c.Timeout,
	}

	if clone.Timeout <= 0 {
		clone.Timeout = 5 * time.Second
	}
	if clone.PoolSize <= 0 {
		clone.PoolSize = 10
	}

	return clone
}

// Addr returns the Redis server address in the format "host:port".
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
