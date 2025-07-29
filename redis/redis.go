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

// RedisCacheConfig holds configuration options for connecting to a Redis instance.
//
// It includes host address, port, authentication credentials, selected DB index,
// connection pool size, and a default timeout (in seconds) for Redis operations.
type RedisCacheConfig struct {
	Host       string // Redis server hostname or IP
	Port       int    // Redis server port
	Password   string // Password for authentication (if required)
	DB         int    // Redis database index (0 by default)
	PoolSize   int    // Maximum number of connections in the pool
	TimeoutSec int    // Timeout for Redis operations in seconds
}

// defaultTimeoutSec defines the default timeout (in seconds) for redis operations.
const defaultTimeoutSec = 5

type RedisCache struct {
	Client     *redis.Client
	TimeoutSec int
	config     *RedisCacheConfig
}

// NewRedisCache initializes a Redis connection using the provided configuration.
//
// It creates a new Redis client, verifies the connection using PING,
// and returns a RedisCache instance. If `TimeoutSec` is zero or negative,
// it falls back to the default timeout.
//
// Returns an error if the connection cannot be established.
//
// Example:
//
//	cache, err := NewRedisCache(&RedisCacheConfig{
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

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	log.Println("Redis connect success")
	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}

	return &RedisCache{
		Client:     rdb,
		TimeoutSec: cf.TimeoutSec,
		config:     cf,
	}, nil
}

func (r *RedisCache) Close() {
	if r.Client != nil {
		_ = r.Client.Close()
	}
}

func (r *RedisCache) GetRDB() (*redis.Client, error) {
	return r.Client, nil
}

func (r *RedisCache) IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

// Setx sets a Redis key with the given value without expiration (persist).
//
// It is a shorthand for calling Set with `expiredTimeSec = -1`.
func (r *RedisCache) Setx(c context.Context, key string, value interface{}) error {
	return r.Set(c, key, value, -1)
}

// Set sets a Redis key to the given value with an optional expiration time (in seconds).
//
// If `expiredTimeSec` is negative, the key is stored without expiration (persist).
// Context is wrapped with an internal timeout (`r.TimeoutSec`).
//
// Example: Set(ctx, "foo", "bar", 60) → key "foo" expires in 60 seconds
func (r *RedisCache) Set(c context.Context, key string, value interface{}, expiredTimeSec int) (err error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, convertValue(value), time.Duration(expiredTimeSec)*time.Second).Err()
}

// SetManyx sets multiple keys in Redis using MSET with no expiration.
//
// This is a faster alternative for bulk writes when expiration is not needed.
// If the input map is empty, it returns immediately.
func (r *RedisCache) SetManyx(c context.Context, data map[string]string) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.MSet(ctx, data).Err()
}

// SetMany sets multiple Redis keys with the same expiration time using a pipeline.
//
// Each key-value pair is inserted using `SET` with the given expiration time (in seconds).
// Internally uses a Redis pipeline for better performance with multiple writes.
//
// Example: SetMany(ctx, map[string]string{"a":"1","b":"2"}, 120)
func (r *RedisCache) SetMany(c context.Context, data map[string]string, expireSec int) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
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
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Array:
		return jsonx.ToJSONBytes(value)
	case reflect.String:
		return value
	default:
		return fmt.Sprint(value)
	}
}

// Get retrieves a value from Redis by key and deserializes it into the given result pointer.
//
// The result must be a pointer to a string, number, or struct. Supports automatic type conversion.
// Uses `reflect` to detect the type of result, and falls back to JSON unmarshal for complex types.
//
// Returns an error if the key does not exist, the type is invalid, or parsing fails.
func (r *RedisCache) Get(c context.Context, key string, result interface{}) (err error) {
	if !validate.IsPtr(result) {
		return errors.New("must be a pointer")
	}

	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
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

// GetString retrieves a Redis value as a plain string.
//
// Returns "" and nil error if the key does not exist.
func (r *RedisCache) GetString(c context.Context, key string) (string, error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
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

// Delete removes the specified key from Redis.
//
// Returns nil if the key does not exist.
func (r *RedisCache) Delete(c context.Context, key string) error {
	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.Del(ctx, key).Err()
}

// GetMany retrieves multiple values from Redis using MGET.
//
// The result is a slice of interface{} values, in the same order as keys.
// Missing keys will be returned as nil in the result slice.
func (r *RedisCache) GetMany(c context.Context, keys []string) ([]interface{}, error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
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

// GetByPrefix scans Redis keys by a given prefix and returns their string values.
//
// Uses SCAN under the hood to avoid blocking. For each matching key, `GetString` is called.
// Returns an error if any key retrieval fails.
func (r *RedisCache) GetByPrefix(c context.Context, prefix string) ([]string, error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
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

// Exists checks whether the given key exists in Redis.
//
// Returns true if the key exists, false otherwise.
func (r *RedisCache) Exists(c context.Context, key string) (bool, error) {
	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return false, err
	}

	count, err := rdb.Exists(ctx, key).Result()
	return count > 0, err
}

// Publish sends a message to a Redis channel (Pub/Sub).
//
// The value will be converted to string using convertValue before sending.
func (r *RedisCache) Publish(c context.Context, channel string, value interface{}) error {
	msg := convertValue(value)
	ctx, cancel := utils.NewCtxTimeout(c, r.TimeoutSec)
	defer cancel()

	rdb, err := r.GetRDB()
	if err != nil {
		return err
	}

	return rdb.Publish(ctx, channel, msg).Err()
}

// Subscribe listens for messages on a given Redis channel and invokes the handler function for each message.
//
// The handler receives the raw message payload as a string.
// The subscription runs in a background goroutine until the context is canceled.
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
