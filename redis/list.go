package redis

import (
	"context"
	"reflect"
	"time"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
)

// listBuilder represents a builder list for Redis list operations with type safety.
type listBuilder[T any] struct {
	cache      *Cache
	key        string
	values     []interface{}
	expiration time.Duration
	start      int64
	end        int64
	setEnd     bool
}

// WithList creates a new list builder list for type T.
func WithList[T any](c *Cache) *listBuilder[T] {
	return &listBuilder[T]{
		cache: c,
	}
}

// Key specifies a single key to operate on for the next execution command.
func (c *listBuilder[T]) Key(k string) *listBuilder[T] {
	c.key = k
	return c
}

// Values specifies multiple values to be stored with the key.
func (c *listBuilder[T]) Values(values interface{}) *listBuilder[T] {
	v := reflect.ValueOf(values)

	if v.Kind() != reflect.Slice {
		c.values = append(c.values, convertValue(values))
		return c
	}

	for i := 0; i < v.Len(); i++ {
		val := v.Index(i).Interface()
		c.values = append(c.values, convertValue(val))
	}

	return c
}

// Expire sets the Time-To-Live (TTL) for the key.
func (c *listBuilder[T]) Expire(d time.Duration) *listBuilder[T] {
	c.expiration = d
	return c
}

// Start sets the start index for range operations.
func (c *listBuilder[T]) Start(start int64) *listBuilder[T] {
	c.start = start
	return c
}

// End sets the end index for range operations.
func (c *listBuilder[T]) End(end int64) *listBuilder[T] {
	c.end = end
	c.setEnd = true
	return c
}

// AddFirst inserts one or more values at the head (left) of the list.
// Returns an error if the key or values are missing, or if the operation fails.
func (c *listBuilder[T]) AddFirst(ctx context.Context) error {
	if str.IsEmpty(c.key) {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	if err := rdb.LPush(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	if c.expiration > 0 {
		_ = rdb.Expire(ct, c.key, c.expiration).Err()
	}
	return nil
}

// Add inserts one or more values at the tail (right) of the list.
// Returns an error if the key or values are missing, or if the operation fails.
func (c *listBuilder[T]) Add(ctx context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	if err := rdb.RPush(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	if c.expiration > 0 {
		_ = rdb.Expire(ct, c.key, c.expiration).Err()
	}
	return nil
}

// PopFront retrieves and removes the first element (head) of the list.
// Returns nil if the list is empty (redis.Nil error).
// Returns an error if the key is missing, or if the operation fails.
func (c *listBuilder[T]) PopFront(ctx context.Context) (*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	val, err := rdb.LPop(ct, c.key).Result()
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

// Pop retrieves and removes the last element (tail) of the list.
// Returns nil if the list is empty (redis.Nil error).
// Returns an error if the key is missing, or if the operation fails.
func (c *listBuilder[T]) Pop(ctx context.Context) (*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	val, err := rdb.RPop(ct, c.key).Result()
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

// GetRange returns a slice of elements between the specified start and stop indexes.
// If end is not set, returns all elements from start to the end of the list.
// Returns an error if the key is missing, or if the operation fails.
func (c *listBuilder[T]) GetRange(ctx context.Context) ([]*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	end := c.end
	if !c.setEnd && end == 0 {
		end = -1 // get all
	}
	vals, err := rdb.LRange(ct, c.key, c.start, end).Result()
	if err != nil {
		return nil, err
	}

	result := make([]*T, 0, len(vals))
	for _, v := range vals {
		t, err := jsonx.FromJSON[T](v)
		if err != nil {
			return nil, err
		}
		result = append(result, &t)
	}
	return result, nil
}

// Get retrieves the element at the specified index from the Redis list.
// Returns nil if the index is out of range.
// Returns an error if the key is missing, or if the operation fails.
func (c *listBuilder[T]) Get(ctx context.Context, index int64) (*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	vals, err := rdb.LRange(ct, c.key, index, index).Result()
	if err != nil {
		return nil, err
	}

	if len(vals) == 0 {
		return nil, nil
	}

	t, err := jsonx.FromJSON[T](vals[0])
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Size returns the number of elements in the list.
// Returns an error if the key is missing, or if the operation fails.
func (c *listBuilder[T]) Size(ctx context.Context) (int64, error) {
	if c.key == "" {
		return 0, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	return rdb.LLen(ct, c.key).Result()
}

// Delete removes the specified key from Redis.
func (c *listBuilder[T]) Delete(ct context.Context) error {
	if str.IsEmpty(c.key) {
		return ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	return rdb.Del(ctx, c.key).Err()
}
