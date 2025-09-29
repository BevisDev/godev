package redis

import (
	"context"
	"encoding/json"
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

const (
	// defaultTimeoutSec defines the default timeout (in seconds) for redis operations.
	defaultTimeoutSec = 60
)

type RedisCache struct {
	client  *redis.Client
	timeout int
	config  *RdConfig
}

// NewRedisCache initializes a Redis connection using the provided configuration.
//
// It creates a new Redis client, verifies the connection using PING,
// and returns a RedisCache instance. If `timeout` is zero or negative,
// it falls back to the default timeout.
//
// Returns an error if the connection cannot be established.
//
// Example:
//
//	cache, err := NewRedisCache(&RdConfig{
//	    Host:     "localhost",
//	    Port:     6379,
//	    Password: "",
//	    DB:       0,
//	    PoolSize: 10,
//	})
//
//	if err != nil {
//	    log.Fatal("Redis init failed:", err)
//	}
func NewRedisCache(cf *RdConfig) (RdExec, error) {
	if cf == nil {
		return nil, errors.New("redis cache config is nil")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cf.Host, cf.Port),
		Password: cf.Password,
		DB:       cf.DB,
		PoolSize: cf.PoolSize,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	log.Println("Redis connect success")
	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}

	return &RedisCache{
		client:  rdb,
		timeout: cf.TimeoutSec,
		config:  cf,
	}, nil
}

func (r *RedisCache) Close() {
	if r.client != nil {
		_ = r.client.Close()
	}
}

func (r *RedisCache) SetTimeout(timeoutSec int) {
	r.timeout = timeoutSec
}

func (r *RedisCache) GetRDB() (*redis.Client, error) {
	return r.client, nil
}

func (r *RedisCache) IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

func (r *RedisCache) Setx(c context.Context, key string, value interface{}) error {
	return r.Set(c, key, value, -1)
}

func (r *RedisCache) Set(c context.Context, key string, value interface{}, expiredTimeSec int) (err error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, convertValue(value), time.Duration(expiredTimeSec)*time.Second).Err()
}

func (r *RedisCache) SetManyx(c context.Context, data map[string]string) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.MSet(ctx, data).Err()
}

func (r *RedisCache) SetMany(c context.Context, data map[string]string, expireSec int) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	pipe := rdb.Pipeline()
	for key, value := range data {
		pipe.Set(ctx, key, value, time.Duration(expireSec)*time.Second)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func convertValue(value interface{}) interface{} {
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

func (r *RedisCache) Get(c context.Context, key string, result interface{}) (err error) {
	if !validate.IsPtr(result) {
		return errors.New("must be a pointer")
	}

	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("redis cache Get panic: %v", r)
		}
	}()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	val, err := rdb.Get(ctx, key).Result()
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
	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return "", err
	}

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if r.IsNil(err) {
			return "", nil
		}
	}
	return val, err
}

func (r *RedisCache) Delete(c context.Context, key string) error {
	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.Del(ctx, key).Err()
}

func (r *RedisCache) GetMany(c context.Context, keys []string) ([]interface{}, error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return nil, err
	}

	vals, err := rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	return vals, nil
}

func (r *RedisCache) GetByPrefix(c context.Context, prefix string) ([]string, error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return nil, err
	}

	var (
		cursor uint64
		result []string
	)
	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, prefix+"*", 0).Result()
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
	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return false, err
	}

	count, err := rdb.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *RedisCache) Publish(c context.Context, channel string, value interface{}) error {
	msg := convertValue(value)
	ctx, cancel := utils.NewCtxTimeout(c, r.timeout)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.Publish(ctx, channel, msg).Err()
}

func (r *RedisCache) Subscribe(ctx context.Context, channel string, handler func(message string)) error {

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	pubsub := rdb.Subscribe(ctx, channel)

	_, err = pubsub.Receive(ctx)
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
