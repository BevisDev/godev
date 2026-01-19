package redis

import (
	"context"
	"time"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/validate"
)

type ChainSet[T any] struct {
	*Chain[T]
}

func WithSet[T any](cache *Cache) ChainSetExec[T] {
	return &ChainSet[T]{
		Chain: withChain[T](cache),
	}
}

func (c *ChainSet[T]) Key(k string) ChainSetExec[T] {
	c.Chain.Key(k)
	return c
}

func (c *ChainSet[T]) Values(values interface{}) ChainSetExec[T] {
	c.Chain.Values(values)
	return c
}

func (c *ChainSet[T]) Expire(d time.Duration) ChainSetExec[T] {
	c.Chain.Expire(d)
	return c
}

func (c *ChainSet[T]) Add(ctx context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.Timeout)
	defer cancel()

	if err := rdb.SAdd(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	if c.expiration > 0 {
		_ = rdb.Expire(ct, c.key, c.expiration).Err()
	}
	return nil
}

func (c *ChainSet[T]) Remove(ctx context.Context) error {
	if c.key == "" {
		return ErrMissingKey
	}
	if validate.IsNilOrEmpty(c.values) {
		return ErrMissingValues
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.Timeout)
	defer cancel()

	if err := rdb.SRem(ct, c.key, c.values...).Err(); err != nil {
		return err
	}

	return nil
}

func (c *ChainSet[T]) Contains(ctx context.Context, val interface{}) (bool, error) {
	if c.key == "" {
		return false, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.Timeout)
	defer cancel()

	return rdb.SIsMember(ct, c.key, val).Result()
}

func (c *ChainSet[T]) GetAll(ctx context.Context) ([]*T, error) {
	if c.key == "" {
		return nil, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.Timeout)
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

func (c *ChainSet[T]) Size(ctx context.Context) (int64, error) {
	if c.key == "" {
		return 0, ErrMissingKey
	}

	rdb := c.GetClient()
	ct, cancel := utils.NewCtxTimeout(ctx, c.Timeout)
	defer cancel()

	return rdb.SCard(ct, c.key).Result()
}

func (c *ChainSet[T]) Delete(ct context.Context) error {
	return c.Chain.Delete(ct)
}
