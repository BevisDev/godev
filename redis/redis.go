// Package redis provides a convenient Redis client for Go with type-safe builder operations.
// It supports basic operations (GET, SET, DELETE, EXISTS), list operations, set operations,
// and pub/sub features with automatic JSON serialization/deserialization.
package redis

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache represents a Redis client connection with configuration.
// It provides methods for interacting with Redis and managing the connection lifecycle.
type Cache struct {
	cf     *Config
	client *redis.Client
}

// New initializes a Redis connection using the provided configuration.
// It creates a new Redis client, verifies the connection using PING,
// and returns a Cache instance. If `timeout` is zero or negative,
// it falls back to the default timeout.
// Returns an error if the connection cannot be established.
func New(cf *Config) (*Cache, error) {
	if cf == nil {
		return nil, errors.New("[redis] config is nil")
	}
	cf.withDefaults()

	var c = &Cache{
		cf: cf,
	}
	rdb, err := c.connect()
	if err != nil {
		return nil, err
	}

	c.client = rdb
	if err := c.Ping(context.Background()); err != nil {
		_ = rdb.Close()
		return nil, err
	}

	log.Println("[redis] connected successfully")
	return c, nil
}

// connect creates a new Redis client with the configured options.
func (r *Cache) connect() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     r.cf.Addr(),
		Password: r.cf.Password,
		DB:       r.cf.DB,
		PoolSize: r.cf.PoolSize,
	})

	return rdb, nil
}

// Ping verifies the connection to Redis by sending a PING command.
// Returns an error if the connection is not available.
func (r *Cache) Ping(ctx context.Context) error {
	if _, err := r.client.Ping(ctx).Result(); err != nil {
		return err
	}
	return nil
}

// Close closes the Redis client connection.
// It is safe to call Close multiple times.
func (r *Cache) Close() {
	if r.client != nil {
		_ = r.client.Close()
	}
}

// GetClient returns the underlying Redis client instance.
// This can be used for advanced operations not covered by the Cache API.
func (r *Cache) GetClient() *redis.Client {
	return r.client
}

// IsNil checks if the error is a Redis nil error (key not found).
// This is useful for distinguishing between "key not found" and other errors.
func (r *Cache) IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

func (r *Cache) SetTimeout(d time.Duration) {
	r.cf.Timeout = d
}
