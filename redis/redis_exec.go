package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RdExec interface {
	// Close is used to release or close the connection to the Redis server
	Close()

	// SetTimeout is used to timeout context for redis operations
	// default 60 seconds
	SetTimeout(timeoutSec int)

	// GetRDB returns a new client instance or a specific read-only connection to Redis.
	GetRDB() (*redis.Client, error)

	// IsNil is to check if an error returned by a Redis command is specifically the "key not found" error.
	IsNil(err error) bool

	// Setx sets a Redis key with the given value without expiration (persist).
	//
	// It is a shorthand for calling Set with `expiredTimeSec = -1`.
	Setx(c context.Context, key string, value interface{}) error

	// SetManyx sets multiple keys in Redis using MSET with no expiration.
	//
	// This is a faster alternative for bulk writes when expiration is not needed.
	// If the input map is empty, it returns immediately.
	SetManyx(c context.Context, data map[string]string) error

	// Set sets a Redis key to the given value with an optional expiration time (in seconds).
	// Example: Set(ctx, "foo", "bar", 60) → key "foo" expires in 60 seconds
	Set(c context.Context, key string, value interface{}, expiredTimeSec int) error

	// SetMany sets multiple Redis keys with the same expiration time using a pipeline.
	// Example: SetMany(ctx, map[string]string{"a":"1","b":"2"}, 120)
	SetMany(c context.Context, data map[string]string, expireSec int) error

	// Get retrieves a value from Redis by key and deserializes it into the given result pointer.
	// Returns an error if the key does not exist, the type is invalid, or parsing fails.
	Get(c context.Context, key string, result interface{}) error

	// GetString retrieves a Redis value as a plain string.
	//
	// Returns "" and nil error if the key does not exist.
	GetString(c context.Context, key string) (string, error)

	// GetMany retrieves multiple values from Redis using MGET.
	//
	// The result is a slice of interface{} values, in the same order as keys.
	// Missing keys will be returned as nil in the result slice.
	GetMany(c context.Context, keys []string) ([]interface{}, error)

	// GetByPrefix scans Redis keys by a given prefix and returns their string values.
	//
	// Uses SCAN under the hood to avoid blocking. For each matching key, `GetString` is called.
	// Returns an error if any key retrieval fails.
	GetByPrefix(c context.Context, prefix string) ([]string, error)

	// Delete removes the specified key from Redis.
	//
	// Returns nil if the key does not exist.
	Delete(c context.Context, key string) error

	// Exists checks whether the given key exists in Redis.
	//
	// Returns true if the key exists, false otherwise.
	Exists(c context.Context, key string) (bool, error)

	// Publish sends a message to a Redis channel (Pub/Sub).
	//
	// The value will be converted to string using convertValue before sending.
	Publish(c context.Context, channel string, value interface{}) error

	// Subscribe listens for messages on a given Redis channel and invokes the handler function for each message.
	//
	// The handler receives the raw message payload as a string.
	// The subscription runs in a background goroutine until the context is canceled.
	Subscribe(ctx context.Context, channel string, handler func(message string)) error
}
