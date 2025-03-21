package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/BevisDev/godev/helper"
	"github.com/redis/go-redis/v9"
)

type RedisCacheConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	PoolSize   int
	TimeoutSec int
}

type RedisCache struct {
	client     *redis.Client
	timeoutSec int
}

func NewRedisCache(ctx context.Context, cf *RedisCacheConfig) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cf.Host, cf.Port),
		Password: cf.Password,
		DB:       cf.DB,
		PoolSize: cf.PoolSize,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	log.Println("Redis connect success")
	return &RedisCache{
		client:     rdb,
		timeoutSec: cf.TimeoutSec,
	}, nil
}

func (r *RedisCache) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *RedisCache) ConvertValue(value interface{}) interface{} {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Struct ||
		v.Kind() == reflect.Map ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Array {
		return helper.ToJSON(value)
	}
	return value
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiredTimeSec int) error {
	ctx, cancel := helper.CreateCtxTimeout(ctx, r.timeoutSec)
	defer cancel()
	err := r.client.Set(ctx, key, r.ConvertValue(value), time.Duration(expiredTimeSec)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) Get(ctx context.Context, key string, result interface{}) error {
	ctx, cancel := helper.CreateCtxTimeout(ctx, r.timeoutSec)
	defer cancel()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	err = helper.JSONToStruct(val, result)
	return err
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	ctx, cancel := helper.CreateCtxTimeout(ctx, r.timeoutSec)
	defer cancel()
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) GetListValueByPrefixKey(ctx context.Context, prefix string) ([]string, error) {
	ctx, cancel := helper.CreateCtxTimeout(ctx, r.timeoutSec)
	defer cancel()
	var (
		cursor uint64
		result []string
	)
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, prefix+"*", 0).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			val, err := r.client.Get(ctx, key).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					result = append(result, val)
					continue
				}
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
