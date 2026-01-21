package redis

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
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
		return nil, errors.New("[redis] config is nil")
	}
	cf.withDefaults()

	var c = &Cache{Config: cf}
	rdb, err := c.connect()
	if err != nil {
		return nil, err
	}

	c.client = rdb
	if err := c.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("[redis] ping failed: %w", err)
	}

	log.Println("[redis] connected successfully")
	return c, nil
}

func (r *Cache) connect() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     r.Addr(),
		Password: r.Password,
		DB:       r.DB,
		PoolSize: r.PoolSize,
	})

	return rdb, nil
}

func (r *Cache) Ping(ctx context.Context) error {
	if _, err := r.client.Ping(ctx).Result(); err != nil {
		return err
	}
	return nil
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
