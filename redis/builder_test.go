package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusApproved  Status = "approved"
)

func (s Status) String() string {
	return string(s)
}

func TestRedisCache_SetAndGet(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf:     &Config{Timeout: 5 * time.Second},
	}
	ctx := context.Background()

	mock.ExpectSet("key", "value", 0).SetVal("OK")
	err := With[string](cache).Key("key").Value("value").Set(ctx)
	require.NoError(t, err)

	mock.ExpectGet("key").SetVal("value")
	result, err := With[string](cache).Key("key").Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "value", *result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Get_KeyNotFound(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	mock.ExpectGet("missing").SetErr(redis.Nil)

	result, err := With[string](cache).Key("missing").Get(ctx)
	require.NoError(t, err)
	assert.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Set_WithTTL(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()
	ttl := 10 * time.Second

	mock.ExpectSet("ttlkey", "val", ttl).SetVal("OK")

	err := With[string](cache).Key("ttlkey").Value("val").Expire(ttl).Set(ctx)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Delete(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
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
	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
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
	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
		},
	}

	assert.True(t, cache.IsNil(redis.Nil))
	assert.False(t, cache.IsNil(errors.New("some error")))
}

func TestRedisCache_Publish(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
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
	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
		},
	}
	ctx := context.Background()

	channel := "user_created"

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

func TestSetIfNotExists(t *testing.T) {
	ctx := context.Background()

	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
		},
	}

	chain := With[string](cache).Key("mykey")

	// if key is not exists return true
	mock.ExpectSetNX("mykey", "hello", 0).SetVal(true)

	chain.Value("hello")
	ok, err := chain.SetIfNotExists(ctx)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())

	// key is existed return false
	mock.ExpectSetNX("mykey", "world", 0).SetVal(false)

	chain.Value("world")
	ok, err = chain.SetIfNotExists(ctx)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSetIfNotExists_EnumValue(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	chain := With[Status](cache)

	chain.Key("statusKey").Value(StatusPending.String())
	mock.ExpectSetNX("statusKey", StatusPending.String(), 0).SetVal(true)
	ok, err := chain.SetIfNotExists(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())

	chain.Key("statusKey").Value(StatusCompleted.String())
	mock.ExpectSetNX("statusKey", StatusCompleted.String(), 0).SetVal(false)
	ok, err = chain.SetIfNotExists(ctx)
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_New_NilConfig(t *testing.T) {
	c, err := New(nil)
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "config is nil")
}
