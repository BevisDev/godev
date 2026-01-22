package redis

import (
	"context"
	"time"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
)

// builder represents a builder for Redis operations with type safety.
// It allows fluent API for building and executing Redis commands.
type builder[T any] struct {
	cache      *Cache
	key        string
	keys       []string
	channel    string
	prefix     string
	value      interface{}
	batches    map[string]interface{}
	expiration time.Duration
}

// With creates a new builder for type T.
func With[T any](c *Cache) *builder[T] {
	return &builder[T]{
		cache: c,
	}
}

// Key specifies a single key to operate on for the next execution command.
func (c *builder[T]) Key(k string) *builder[T] {
	c.key = k
	return c
}

// Keys specifies multiple keys for bulk operations.
func (c *builder[T]) Keys(keys ...string) *builder[T] {
	c.keys = keys
	return c
}

// Value specifies the single value to be stored with the key.
func (c *builder[T]) Value(v interface{}) *builder[T] {
	c.value = convertValue(v)
	return c
}

// Put adds a key-value pair to the batch for SetMany operation.
func (c *builder[T]) Put(k string, v interface{}) *builder[T] {
	if c.batches == nil {
		c.batches = make(map[string]interface{})
	}
	c.batches[k] = convertValue(v)
	return c
}

// Batch sets multiple key-value pairs for SetMany operation.
func (c *builder[T]) Batch(b map[string]interface{}) *builder[T] {
	if c.batches == nil {
		c.batches = make(map[string]interface{})
	}

	for k, v := range b {
		c.batches[k] = convertValue(v)
	}
	return c
}

// Expire sets the Time-To-Live (TTL) for the key.
func (c *builder[T]) Expire(d time.Duration) *builder[T] {
	c.expiration = d
	return c
}

// Channel specifies the channel to be used for Pub/Sub operations.
func (c *builder[T]) Channel(channel string) *builder[T] {
	c.channel = channel
	return c
}

// Prefix sets a prefix to be automatically prepended to all subsequent keys in the builder.
func (c *builder[T]) Prefix(prefix string) *builder[T] {
	c.prefix = prefix
	return c
}

// Set sets a Redis key to the given value with an optional expiration time.
// Returns an error if the key or value is missing, or if the operation fails.
func (c *builder[T]) Set(ct context.Context) error {
	if str.IsEmpty(c.key) {
		return ErrMissingKey
	}
	if c.value == nil {
		return ErrMissingValue
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	return rdb.Set(ctx, c.key, c.value, c.expiration).Err()
}

// SetIfNotExists sets the value of the key only if the key does not already exist.
// Returns true if the value was set, false if the key already exists.
// Returns an error if the key or value is missing, or if the operation fails.
func (c *builder[T]) SetIfNotExists(ct context.Context) (bool, error) {
	if str.IsEmpty(c.key) {
		return false, ErrMissingKey
	}
	if c.value == nil {
		return false, ErrMissingValue
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	return rdb.SetNX(ctx, c.key, c.value, c.expiration).Result()
}

// SetMany sets multiple Redis keys with the same expiration time using a pipeline.
// Returns an error if batch data is missing, or if the operation fails.
func (c *builder[T]) SetMany(ct context.Context) error {
	if validate.IsNilOrEmpty(c.batches) {
		return ErrMissingPushOrBatch
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	pipe := rdb.Pipeline()
	for key, value := range c.batches {
		pipe.Set(ctx, key, value, c.expiration)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (c *builder[T]) Get(ct context.Context) (*T, error) {
	if str.IsEmpty(c.key) {
		return nil, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	val, err := rdb.Get(ctx, c.key).Result()
	if err != nil {
		if c.cache.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}

	t, err := jsonx.FromJSON[T](val)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *builder[T]) GetMany(ct context.Context) ([]*T, error) {
	if len(c.keys) <= 0 {
		return nil, ErrMissingKeys
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	vals, err := rdb.MGet(ctx, c.keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make([]*T, 0, len(vals))
	for _, v := range vals {
		if v == nil {
			result = append(result, nil)
			continue
		}

		switch val := v.(type) {
		case string:
			t, err := jsonx.FromJSON[T](val)
			if err != nil {
				return nil, err
			}

			result = append(result, &t)
		case []byte:
			t, err := jsonx.FromJSONBytes[T](val)
			if err != nil {
				return nil, err
			}

			result = append(result, &t)
		default:
			continue
		}
	}

	return result, nil
}

func (c *builder[T]) GetByPrefix(ct context.Context) ([]*T, error) {
	if str.IsEmpty(c.prefix) {
		return nil, ErrMissingPrefix
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	var (
		cursor uint64
		result []*T
	)
	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, c.prefix+"*", 0).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			clone := c
			clone.key = key

			val, err := clone.Get(ctx)
			if err != nil {
				return nil, err
			}
			if val != nil {
				result = append(result, val)
			}
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return result, nil
}

func (c *builder[T]) Delete(ct context.Context) error {
	if str.IsEmpty(c.key) {
		return ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	return rdb.Del(ctx, c.key).Err()
}

func (c *builder[T]) Exists(ct context.Context) (bool, error) {
	if str.IsEmpty(c.key) {
		return false, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	count, err := rdb.Exists(ctx, c.key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (c *builder[T]) Publish(ct context.Context) error {
	if str.IsEmpty(c.channel) {
		return ErrMissingChannel
	}
	if c.value == nil {
		return ErrMissingValue
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	return rdb.Publish(ctx, c.channel, c.value).Err()
}

// Subscribe listens for messages on a given Redis channel and invokes the handler function for each message.
// The handler receives the raw message payload as a string.
// The subscription runs in a background goroutine until the context is canceled.
// Returns an error if the channel is missing, or if the subscription fails.
func (c *builder[T]) Subscribe(ctx context.Context, handler func(msg string)) error {
	if str.IsEmpty(c.channel) {
		return ErrMissingChannel
	}

	rdb := c.cache.GetClient()
	pubsub := rdb.Subscribe(ctx, c.channel)
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return err
	}

	ch := pubsub.Channel()
	go func() {
		defer pubsub.Close()
		for {
			select {
			case msg := <-ch:
				if msg == nil {
					continue
				}

				handler(msg.Payload)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
