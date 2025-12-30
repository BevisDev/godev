package redis

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// Config holds configuration options for connecting to a Redis instance.
//
// It includes host address, port, authentication credentials, selected DB index,
// connection pool size, and a default timeout (in seconds) for Redis operations.
type Config struct {
	Host       string // Redis server hostname or IP
	Port       int    // Redis server port
	Password   string // Password for authentication (if required)
	DB         int    // Redis database index (0 by default)
	PoolSize   int    // Maximum number of connections in the pool
	TimeoutSec int    // timeout for Redis operations in seconds
}

const (
	// defaultTimeoutSec defines the default timeout (in seconds) for redis operations.
	defaultTimeoutSec = 60
	defaultPoolSize   = 10
)

type Cache struct {
	*Config
	client *redis.Client
}

// NewCache initializes a Redis connection using the provided configuration.
// It creates a new Redis client, verifies the connection using PING,
// and returns a Cache instance. If `timeout` is zero or negative,
// it falls back to the default timeout.
// Returns an error if the connection cannot be established.
func NewCache(cf *Config) (*Cache, error) {
	if cf == nil {
		return nil, errors.New("config is nil")
	}

	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}
	if cf.PoolSize <= 0 {
		cf.PoolSize = defaultPoolSize
	}

	var c = &Cache{Config: cf}
	rdb, err := c.connect()
	if err != nil {
		return nil, err
	}

	c.client = rdb
	return c, nil
}

func (r *Cache) connect() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			r.Host, r.Port),
		Password: r.Password,
		DB:       r.DB,
		PoolSize: r.PoolSize,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	log.Printf("connect redis %d success", r.DB)
	return rdb, nil
}

func (r *Cache) Close() {
	if r.client != nil {
		_ = r.client.Close()
	}
}

func (r *Cache) GetClient() *redis.Client {
	return r.client
}

func (r *Cache) IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
