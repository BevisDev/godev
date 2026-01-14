package redis

import (
	"context"
	"fmt"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/validate"
)

type ChainList[T any] struct {
	*Chain[T]
	start  int64
	end    int64
	setEnd bool
}

func WithList[T any](cache *Cache) ChainListExec[T] {
	return &ChainList[T]{
		Chain: withChain[T](cache),
	}
}

func (c *ChainList[T]) Key(k string) ChainListExec[T] {
	c.Chain.Key(k)
	return c
}

func (c *ChainList[T]) Values(values interface{}) ChainListExec[T] {
	c.Chain.Values(values)
	return c
}

func (c *ChainList[T]) Expire(n int, unit string) ChainListExec[T] {
	c.Chain.Expire(n, unit)
	return c
}

func (c *ChainList[T]) Start(start int64) ChainListExec[T] {
	c.start = start
	return c
}

func (c *ChainList[T]) End(end int64) ChainListExec[T] {
	c.end = end
	c.setEnd = true
	return c
}

func (c *ChainList[T]) AddFirst(ctx context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.TimeoutSec)
	defer cancel()

	if err := rdb.LPush(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	if c.expiration > 0 {
		_ = rdb.Expire(ct, c.key, c.expiration).Err()
	}
	return nil
}

func (c *ChainList[T]) Add(ctx context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.TimeoutSec)
	defer cancel()

	if err := rdb.RPush(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	if c.expiration > 0 {
		_ = rdb.Expire(ct, c.key, c.expiration).Err()
	}
	return nil
}

func (c *ChainList[T]) PopFront(ctx context.Context) (*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.TimeoutSec)
	defer cancel()

	val, err := rdb.LPop(ct, c.key).Result()
	if err != nil {
		if c.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}

	var t T
	if _, ok := any(t).(string); ok {
		t = any(val).(T)
	} else if err := jsonx.ToStruct(val, &t); err != nil {
		return nil, fmt.Errorf("parse to %T failed: %w", t, err)
	}
	return &t, nil
}

func (c *ChainList[T]) Pop(ctx context.Context) (*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.TimeoutSec)
	defer cancel()

	val, err := rdb.RPop(ct, c.key).Result()
	if err != nil {
		if c.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}

	var t T
	if _, ok := any(t).(string); ok {
		t = any(val).(T)
	} else if err := jsonx.ToStruct(val, &t); err != nil {
		return nil, fmt.Errorf("parse to %T failed: %w", t, err)
	}
	return &t, nil
}

func (c *ChainList[T]) GetRange(ctx context.Context) ([]*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.TimeoutSec)
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
		var t T
		if _, ok := any(t).(string); ok {
			t = any(v).(T)
		} else if err := jsonx.ToStruct(v, &t); err != nil {
			return nil, fmt.Errorf("parse to %T failed: %w", t, err)
		}
		result = append(result, &t)
	}
	return result, nil
}

func (c *ChainList[T]) Get(ctx context.Context, index int64) (*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.TimeoutSec)
	defer cancel()

	vals, err := rdb.LRange(ct, c.key, index, index).Result()
	if err != nil {
		return nil, err
	}

	if len(vals) == 0 {
		return nil, nil
	}

	var t T
	if _, ok := any(t).(string); ok {
		t = any(vals[0]).(T)
	} else {
		if err := jsonx.ToStruct(vals[0], &t); err != nil {
			return nil, fmt.Errorf("parse to %T failed: %w", t, err)
		}
	}

	return &t, nil
}

func (c *ChainList[T]) Size(ctx context.Context) (int64, error) {
	if c.key == "" {
		return 0, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.TimeoutSec)
	defer cancel()

	return rdb.LLen(ct, c.key).Result()
}

func (c *ChainList[T]) Delete(ct context.Context) error {
	return c.Chain.Delete(ct)
}
