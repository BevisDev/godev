package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/utils/jsonx"
	"github.com/BevisDev/godev/utils/validate"
	"log"
	"reflect"
	"strconv"
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
	if cf == nil {
		return nil, errors.New("redis cache config is nil")
	}
	
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

func (r *RedisCache) SetManyx(c context.Context, data map[string]string) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	return r.Client.MSet(ctx, data).Err()
}

func (r *RedisCache) SetMany(c context.Context, data map[string]string, expireSec int) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	pipe := r.Client.Pipeline()
	for key, value := range data {
		pipe.Set(ctx, key, value, time.Duration(expireSec)*time.Second)
	}

	_, err := pipe.Exec(ctx)
	return err
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
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("redis cache Get panic: %v", r)
		}
	}()

	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return
	}

	v := reflect.ValueOf(result).Elem()
	switch v.Kind() {
	case reflect.String:
		v.SetString(val)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, parseErr := strconv.ParseInt(val, 10, 64)
		if parseErr != nil {
			return parseErr
		}
		v.SetInt(parsed)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, parseErr := strconv.ParseUint(val, 10, 64)
		if parseErr != nil {
			return parseErr
		}
		v.SetUint(parsed)
		return nil
	case reflect.Float32, reflect.Float64:
		parsed, parseErr := strconv.ParseFloat(val, 64)
		if parseErr != nil {
			return parseErr
		}
		v.SetFloat(parsed)
		return nil
	default:
		return jsonx.ToStruct(val, result)
	}
}

func (r *RedisCache) GetString(c context.Context, key string) (string, error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if r.IsNil(err) {
			return "", nil
		}
	}
	return val, err
}

func (r *RedisCache) Delete(c context.Context, key string) (err error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()
	return r.Client.Del(ctx, key).Err()
}

func (r *RedisCache) GetMany(c context.Context, keys []string) ([]interface{}, error) {
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

func (r *RedisCache) Exists(c context.Context, key string) (bool, error) {
	ctx, cancel := utils.CreateCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	count, err := r.Client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *RedisCache) Publish(ctx context.Context, channel string, value interface{}) error {
	msg := convertValue(value)
	return r.Client.Publish(ctx, channel, msg).Err()
}

func (r *RedisCache) Subscribe(ctx context.Context, channel string, handler func(message string)) error {
	pubsub := r.Client.Subscribe(ctx, channel)

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
