package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/validate"
	"log"
	"reflect"
	"time"

	"github.com/BevisDev/godev/utils"
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

var defaultTimeoutSec = 30

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
	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}
	return &RedisCache{rdb, cf.TimeoutSec}, nil
}

func (r *RedisCache) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}

func (r *RedisCache) IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

func (r *RedisCache) Setx(c context.Context, key string, value interface{}) error {
	return r.Set(c, key, value, -1)
}

func (r *RedisCache) Set(c context.Context, key string, value interface{}, expiredTimeSec int) (err error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	return r.Client.Set(ctx, key, convertValue(value), time.Duration(expiredTimeSec)*time.Second).Err()
}

func convertValue(value interface{}) interface{} {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Struct ||
		v.Kind() == reflect.Map ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Array {
		return jsonx.ToJSONBytes(value)
	}
	return value
}

func (r *RedisCache) Get(c context.Context, key string, result interface{}) (err error) {
	if !validate.IsPtr(result) {
		return errors.New("must be a pointer")
	}
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return
	}
	return jsonx.ToStruct(val, result)
}

func (r *RedisCache) GetString(c context.Context, key string) (string, error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if r.IsNil(err) {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func (r *RedisCache) Delete(c context.Context, key string) (err error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	return r.Client.Del(ctx, key).Err()
}

func (r *RedisCache) SetBatch(c context.Context, args map[string]string) (err error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	return r.Client.MSet(ctx, args).Err()
}

func (r *RedisCache) GetBatch(c context.Context, keys []string) ([]interface{}, error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	vals, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	return vals, nil
}

func (r *RedisCache) GetByPrefix(c context.Context, prefix string) ([]string, error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
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
			val, err := r.GetString(ctx, key)
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
