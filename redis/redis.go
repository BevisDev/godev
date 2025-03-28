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
	Client     *redis.Client
	TimeoutSec int
}

func NewRedisCache(cf *RedisCacheConfig) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cf.Host, cf.Port),
		Password: cf.Password,
		DB:       cf.DB,
		PoolSize: cf.PoolSize,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	log.Println("Redis connect success")

	return &RedisCache{
		Client:     rdb,
		TimeoutSec: cf.TimeoutSec,
	}, nil
}

func (r *RedisCache) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}

func (r *RedisCache) convertValue(value interface{}) interface{} {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Struct ||
		v.Kind() == reflect.Map ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Array {
		return helper.ToJSONBytes(value)
	}
	return value
}

func (r *RedisCache) Set(c context.Context, key string, value interface{}, expiredTimeSec int) error {
	ctx, cancel := helper.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	err := r.Client.Set(ctx, key, r.convertValue(value), time.Duration(expiredTimeSec)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) Get(c context.Context, key string, result interface{}) error {
	ctx, cancel := helper.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	err = helper.JSONToStruct(val, result)
	return err
}

func (r *RedisCache) Delete(c context.Context, key string) error {
	ctx, cancel := helper.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCache) GetListValueByPrefixKey(c context.Context, prefix string) ([]string, error) {
	ctx, cancel := helper.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	var (
		cursor uint64
		result []string
	)
	for {
		keys, nextCursor, err := r.Client.Scan(ctx, cursor, prefix+"*", 0).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			val, err := r.Client.Get(ctx, key).Result()
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
