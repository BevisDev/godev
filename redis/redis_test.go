package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestRedisCache_SetAndGet(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	mock.ExpectSet("key", "value", 0).SetVal("OK")
	err := cache.Setx(ctx, "key", "value")
	assert.NoError(t, err)

	mock.ExpectGet("key").SetVal("value")
	var result string
	err = cache.Get(ctx, "key", &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Delete(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	mock.ExpectDel("key").SetVal(1)
	err := cache.Delete(ctx, "key")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_SetManyx(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	mock.ExpectMSet(data).SetVal("OK")

	err := cache.SetManyx(ctx, data)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_SetManyx_GetMany(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	batch := map[string]string{"key1": "value1", "key2": "value2"}
	mock.ExpectMSet(batch).SetVal("OK")
	err := cache.SetManyx(ctx, batch)
	assert.NoError(t, err)

	mock.ExpectMGet("key1", "key2").SetVal([]interface{}{"value1", "value2"})
	vals, err := cache.GetMany(ctx, []string{"key1", "key2"})
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"value1", "value2"}, vals)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_SetMany(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	expireSec := 60

	mock.ExpectSet("key1", "value1", time.Duration(expireSec)*time.Second).SetVal("OK")
	mock.ExpectSet("key2", "value2", time.Duration(expireSec)*time.Second).SetVal("OK")

	err := cache.SetMany(ctx, data, expireSec)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_GetByPrefix(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	mock.ExpectScan(0, "prefix*", int64(0)).SetVal([]string{"prefix1", "prefix2"}, 0)
	mock.ExpectGet("prefix1").SetVal("value1")
	mock.ExpectGet("prefix2").SetVal("value2")

	vals, err := cache.GetByPrefix(ctx, "prefix")
	assert.NoError(t, err)
	assert.Equal(t, []string{"value1", "value2"}, vals)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Get_NotPointer(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	err := cache.Get(ctx, "key", "not-pointer")
	assert.EqualError(t, err, "must be a pointer")
}

func TestRedisCache_IsNil(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &RedisCache{Client: rdb, TimeoutSec: 5}

	assert.True(t, cache.IsNil(redis.Nil))
	assert.False(t, cache.IsNil(errors.New("some error")))
}

func TestRedisCache_Publish(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	redisCache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	channel := "test_channel"
	message := "hello world"

	mock.ExpectPublish(channel, message).SetVal(1)

	err := redisCache.Publish(ctx, channel, message)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Publish_JSON(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	redisCache := &RedisCache{Client: rdb, TimeoutSec: 5}
	ctx := context.Background()

	channel := "user_created"

	// Struct để publish
	user := User{
		ID:   1,
		Name: "Alice",
	}

	jsonBytes, err := json.Marshal(user)
	assert.NoError(t, err)

	mock.ExpectPublish(channel, jsonBytes).SetVal(1)

	err = redisCache.Publish(ctx, channel, user)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
