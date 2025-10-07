package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/validate"
	"reflect"
	"time"
)

type Chain[T any] struct {
	*RedisCache
	key        string
	keys       []string
	channel    string
	prefix     string
	value      interface{}
	values     []interface{} // for list or set
	batches    map[string]interface{}
	expiration time.Duration
}

func With[T any](cache *RedisCache) ChainExec[T] {
	return &Chain[T]{
		RedisCache: cache,
	}
}

func withChain[T any](cache *RedisCache) *Chain[T] {
	return &Chain[T]{
		RedisCache: cache,
	}
}

func (c *Chain[T]) Key(k string) ChainExec[T] {
	c.key = k
	return c
}

func (c *Chain[T]) Keys(keys ...string) ChainExec[T] {
	c.keys = keys
	return c
}

func (c *Chain[T]) Value(v interface{}) ChainExec[T] {
	c.value = c.convertValue(v)
	return c
}

func (c *Chain[T]) Values(values interface{}) ChainExec[T] {
	v := reflect.ValueOf(values)

	if v.Kind() != reflect.Slice {
		c.values = append(c.values, c.convertValue(values))
		return c
	}

	for i := 0; i < v.Len(); i++ {
		val := v.Index(i).Interface()
		c.values = append(c.values, c.convertValue(val))
	}

	return c
}

func (c *Chain[T]) Put(k string, v interface{}) ChainExec[T] {
	if c.batches == nil {
		c.batches = make(map[string]interface{})
	}
	c.batches[k] = c.convertValue(v)
	return c
}

func (c *Chain[T]) Batch(b map[string]interface{}) ChainExec[T] {
	if c.batches == nil {
		c.batches = make(map[string]interface{})
	}

	for k, v := range b {
		c.batches[k] = c.convertValue(v)
	}
	return c
}

func (c *Chain[T]) Expire(n int, unit string) ChainExec[T] {
	switch unit {
	case "s", "sec", "second", "seconds":
		c.expiration = time.Duration(n) * time.Second
	case "m", "min", "minute", "minutes":
		c.expiration = time.Duration(n) * time.Minute
	case "h", "hr", "hour", "hours":
		c.expiration = time.Duration(n) * time.Hour
	default:
		c.expiration = time.Duration(n) * time.Second
	}

	return c
}

func (c *Chain[T]) convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return v
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return fmt.Sprint(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return b
	}
}

func (c *Chain[T]) Channel(channel string) ChainExec[T] {
	c.channel = channel
	return c
}

func (c *Chain[T]) Prefix(prefix string) ChainExec[T] {
	c.prefix = prefix
	return c
}

func (c *Chain[T]) Set(ct context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if c.value == nil {
		return ErrMissingValue
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
	defer cancel()

	return rdb.Set(ctx, c.key, c.value, c.expiration).Err()
}

func (c *Chain[T]) SetIfNotExists(ct context.Context) (bool, error) {
	if c.key == "" {
		return false, ErrMissingKey
	}
	if c.value == nil {
		return false, ErrMissingValue
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
	defer cancel()

	return rdb.SetNX(ctx, c.key, c.value, c.expiration).Result()
}

func (c *Chain[T]) SetMany(ct context.Context) error {
	if validate.IsNilOrEmpty(c.batches) {
		return ErrMissingPushOrBatch
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
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

func (c *Chain[T]) Get(ct context.Context) (*T, error) {
	var err error
	if c.key == "" {
		return nil, ErrMissingKey
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
	defer cancel()

	val, err := rdb.Get(ctx, c.key).Result()
	if err != nil {
		if c.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}

	var t T
	if _, ok := any(t).(string); ok {
		t = any(val).(T)
	} else {
		if err := jsonx.ToStruct(val, &t); err != nil {
			return nil, fmt.Errorf("parse to %T failed: %w", t, err)
		}
	}

	return &t, nil
}

func (c *Chain[T]) GetMany(ct context.Context) ([]*T, error) {
	if validate.IsNilOrEmpty(c.keys) {
		return nil, ErrMissingKeys
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
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

		var strVal string
		switch val := v.(type) {
		case string:
			strVal = val
		case []byte:
			strVal = string(val)
		default:
			continue
		}

		var t T
		if _, ok := any(t).(string); ok {
			t = any(strVal).(T)
		} else {
			if err := jsonx.ToStruct(strVal, &t); err != nil {
				return nil, fmt.Errorf("parse to %T failed: %w", t, err)
			}
		}

		result = append(result, &t)
	}

	return result, nil
}

func (c *Chain[T]) GetByPrefix(ct context.Context) ([]*T, error) {
	if c.prefix == "" {
		return nil, ErrMissingPrefix
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
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
			var clone = c
			clone.key = key

			val, err := clone.Get(ctx)
			if err != nil {
				return nil, err
			}
			result = append(result, val)
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return result, nil
}

func (c *Chain[T]) Delete(ct context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
	defer cancel()

	return rdb.Del(ctx, c.key).Err()
}

func (c *Chain[T]) Exists(ct context.Context) (bool, error) {
	if c.key == "" {
		return false, ErrMissingKey
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
	defer cancel()

	count, err := rdb.Exists(ctx, c.key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (c *Chain[T]) Publish(ct context.Context) error {
	if c.channel == "" {
		return ErrMissingChannel
	}
	if c.value == nil {
		return ErrMissingValue
	}

	rdb := c.GetRDB()
	ctx, cancel := utils.NewCtxTimeout(ct, c.TimeoutSec)
	defer cancel()

	return rdb.Publish(ctx, c.channel, c.value).Err()
}

func (c *Chain[T]) Subscribe(ctx context.Context, handler func(msg string)) error {
	if c.channel == "" {
		return ErrMissingChannel
	}

	rdb := c.GetRDB()
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
