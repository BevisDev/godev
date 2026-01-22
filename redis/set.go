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

// setBuilder represents a builder set for Redis set operations with type safety.
type setBuilder[T any] struct {
	cache      *Cache
	key        string
	values     []interface{}
	expiration time.Duration
}

// WithSet creates a new set builder set for type T.
func WithSet[T any](c *Cache) *setBuilder[T] {
	return &setBuilder[T]{
		cache: c,
	}
}

// Key specifies a single key to operate on for the next execution command.
func (c *setBuilder[T]) Key(k string) *setBuilder[T] {
	c.key = k
	return c
}

// Values specifies multiple values to be stored with the key.
func (c *setBuilder[T]) Values(values interface{}) *setBuilder[T] {
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
func (c *setBuilder[T]) Expire(d time.Duration) *setBuilder[T] {
	c.expiration = d
	return c
}

// Add adds one or more members to the set.
// Returns an error if the key or values are missing, or if the operation fails.
func (c *setBuilder[T]) Add(ctx context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	if err := rdb.SAdd(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	if c.expiration > 0 {
		_ = rdb.Expire(ct, c.key, c.expiration).Err()
	}
	return nil
}

// Remove removes one or more members from the set.
// Returns an error if the key or values are missing, or if the operation fails.
func (c *setBuilder[T]) Remove(ctx context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	if err := rdb.SRem(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	return nil
}

// Contains checks if a value exists in the set.
// Returns an error if the key is missing, or if the operation fails.
func (c *setBuilder[T]) Contains(ctx context.Context, val interface{}) (bool, error) {
	if c.key == "" {
		return false, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	return rdb.SIsMember(ct, c.key, val).Result()
}

// GetAll returns all members of the set.
// Returns an error if the key is missing, or if the operation fails.
func (c *setBuilder[T]) GetAll(ctx context.Context) ([]*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	res, err := rdb.SMembers(ct, c.key).Result()
	if err != nil {
		return nil, err
	}

	result := make([]*T, 0, len(res))
	for _, v := range res {
		t, err := jsonx.FromJSON[T](v)
		if err != nil {
			return nil, err
		}
		result = append(result, &t)
	}

	return result, nil
}

// Size returns the number of elements in the set.
// Returns an error if the key is missing, or if the operation fails.
func (c *setBuilder[T]) Size(ctx context.Context) (int64, error) {
	if c.key == "" {
		return 0, ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.cache.cf.Timeout)
	defer cancel()

	return rdb.SCard(ct, c.key).Result()
}

// Delete removes the specified key from Redis.
func (c *setBuilder[T]) Delete(ct context.Context) error {
	if str.IsEmpty(c.key) {
		return ErrMissingKey
	}

	rdb := c.cache.GetClient()
	ctx, cancel := utils.NewCtxTimeout(ct, c.cache.cf.Timeout)
	defer cancel()

	return rdb.Del(ctx, c.key).Err()
}
