package providers

import (
	"context"

	"github.com/BevisDev/godev/redis"
)

type RedisProvider struct {
	cfg   *redis.Config
	cache *redis.Cache
}

func NewRedisProvider(cfg *redis.Config) *RedisProvider {
	return &RedisProvider{cfg: cfg}
}

func (p *RedisProvider) Init(ctx context.Context) error {
	_ = ctx
	cache, err := redis.New(p.cfg)
	if err != nil {
		return err
	}
	p.cache = cache
	return nil
}

func (p *RedisProvider) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *RedisProvider) Stop(ctx context.Context) error {
	_ = ctx
	if p.cache != nil {
		p.cache.Close()
	}
	return nil
}

func (p *RedisProvider) Cache() *redis.Cache {
	return p.cache
}
