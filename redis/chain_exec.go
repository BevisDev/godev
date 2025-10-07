package redis

import (
	"context"
)

type ChainExec[T any] interface {
	// Key specifies a single key to operate on for the next execution command
	Key(k string) ChainExec[T]

	// Keys specifies multiple keys for bulk operations
	Keys(keys ...string) ChainExec[T]

	// Value specifies single the value to be stored with the key
	Value(v interface{}) ChainExec[T]

	// Values specifies multiple values to be stored with the key
	Values(values interface{}) ChainExec[T]

	// Expire sets the Time-To-Live (TTL) for the key.
	// n is the time duration, and unit specifies the scale (e.g., "s" for seconds).
	Expire(n int, unit string) ChainExec[T]

	// Put executes a direct Set operation for a single Key-Value pair,
	Put(k string, v interface{}) ChainExec[T]

	// Batch executes a bulk Set operation (like MSet) using the provided map of Key-Value pairs.
	Batch(b map[string]interface{}) ChainExec[T]

	// Channel specifies the channel to be used for Pub/Sub
	Channel(channel string) ChainExec[T]

	// Prefix sets a prefix to be automatically prepended to all subsequent keys in the chain.
	Prefix(prefix string) ChainExec[T]

	// Set sets a Redis key to the given value with an optional expiration time (in seconds).
	Set(c context.Context) error

	// SetIfNotExists sets the value of the key only if the key does not already exist.
	// Returns true if the value was set, false if the key already exists.
	SetIfNotExists(ct context.Context) (bool, error)

	// SetMany sets multiple Redis keys with the same expiration time
	SetMany(ct context.Context) error

	// Get retrieves a value from Redis by key
	Get(ct context.Context) (*T, error)

	// GetMany retrieves multiple values from Redis
	// The result is a slice of interface{} values, in the same order as keys.
	// Missing keys will be returned as nil in the result slice.
	GetMany(c context.Context) ([]*T, error)

	// GetByPrefix scans Redis keys by a given prefix
	// Returns an error if any key retrieval fails.
	GetByPrefix(c context.Context) ([]*T, error)

	// Delete removes the specified key from Redis.
	Delete(ct context.Context) error

	// Exists checks whether the given key exists in Redis.
	// Returns true if the key exists, false otherwise.
	Exists(c context.Context) (bool, error)

	// Publish sends a message to a Redis channel (Pub/Sub).
	Publish(ct context.Context) error

	// Subscribe listens for messages on a given Redis channel and invokes the handler function for each message.
	// The handler receives the raw message payload as a string.
	// The subscription runs in a background goroutine until the context is canceled.
	Subscribe(ctx context.Context, handler func(message string)) error
}
