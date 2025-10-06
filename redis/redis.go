package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
)

const (
	// defaultTimeoutSec defines the default timeout (in seconds) for redis operations.
	defaultTimeoutSec = 60
	defaultPoolSize   = 10
)

type RedisCache struct {
	*Config
	client *redis.Client
}

// New initializes a Redis connection using the provided configuration.
//
// It creates a new Redis client, verifies the connection using PING,
// and returns a RedisCache instance. If `timeout` is zero or negative,
// it falls back to the default timeout.
// Returns an error if the connection cannot be established.
func New(cf *Config) (*RedisCache, error) {
	if cf == nil {
		return nil, errors.New("config is nil")
	}

	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}
	if cf.PoolSize <= 0 {
		cf.PoolSize = defaultPoolSize
	}

	var cache = &RedisCache{Config: cf}
	rdb, err := cache.connect()
	if err != nil {
		return nil, err
	}
	cache.client = rdb

	return cache, nil
}

func (r *RedisCache) connect() (*redis.Client, error) {
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

func (r *RedisCache) Close() {
	if r.client != nil {
		_ = r.client.Close()
		r.client = nil
	}
}

func (r *RedisCache) GetRDB() *redis.Client {
	return r.client
}

func (r *RedisCache) IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
