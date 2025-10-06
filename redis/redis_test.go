package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestRedisCache_SetAndGet(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{
		client: rdb,
		Config: &Config{
			TimeoutSec: 5,
		},
	}
	ctx := context.Background()

	mock.ExpectSet("key", "value", 0).SetVal("OK")
	err := With[string](cache).Key("key").Value("value").Set(ctx)
	assert.NoError(t, err)

	mock.ExpectGet("key").SetVal("value")
	result, err := With[string](cache).Key("key").Get(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "value", *result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Delete(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{
		client: rdb,
		Config: &Config{
			TimeoutSec: 5,
		},
	}
	ctx := context.Background()

	mock.ExpectDel("key").SetVal(1)
	err := With[string](cache).Key("key").Delete(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_GetByPrefix(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{
		client: rdb,
		Config: &Config{
			TimeoutSec: 5,
		},
	}
	ctx := context.Background()

	mock.ExpectScan(0, "prefix*", int64(0)).SetVal([]string{"prefix1", "prefix2"}, 0)
	mock.ExpectGet("prefix1").SetVal("value1")
	mock.ExpectGet("prefix2").SetVal("value2")

	vals, err := With[string](cache).
		Prefix("prefix").
		GetByPrefix(ctx)

	got := make([]string, 0, len(vals))
	for _, v := range vals {
		if v != nil {
			got = append(got, *v)
		}
	}

	assert.NoError(t, err)
	assert.Equal(t, []string{"value1", "value2"}, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_IsNil(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &RedisCache{
		client: rdb,
		Config: &Config{
			TimeoutSec: 5,
		},
	}

	assert.True(t, cache.IsNil(redis.Nil))
	assert.False(t, cache.IsNil(errors.New("some error")))
}

func TestRedisCache_Publish(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{
		client: rdb,
		Config: &Config{
			TimeoutSec: 5,
		},
	}
	ctx := context.Background()

	channel := "test_channel"
	message := "hello world"

	mock.ExpectPublish(channel, message).SetVal(1)

	err := With[string](cache).
		Channel(channel).
		Value(message).
		Publish(ctx)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Publish_JSON(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &RedisCache{
		client: rdb,
		Config: &Config{
			TimeoutSec: 5,
		},
	}
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

	err = With[User](cache).
		Channel(channel).
		Value(&user).
		Publish(ctx)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
