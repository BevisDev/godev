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

	mock.ExpectSet("key", []byte("value"), 0).SetVal("OK")
	err := With[string](cache).Key("key").Value("value").Set(ctx)
	require.NoError(t, err)

	mock.ExpectGet("key").SetVal("value")
	result, err := With[string](cache).Key("key").Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, "value", result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Get_KeyNotFound(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	mock.ExpectGet("missing").SetErr(redis.Nil)

	result, err := With[string](cache).Key("missing").Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Set_WithTTL(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()
	ttl := 10 * time.Second

	mock.ExpectSet("ttlkey", []byte("val"), ttl).SetVal("OK")

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

	assert.NoError(t, err)
	assert.Equal(t, []string{"value1", "value2"}, vals)
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

	mock.ExpectPublish(channel, []byte(message)).SetVal(1)

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

func TestRedisCache_Publish_MissingChannel(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()
	err := With[string](cache).Value("msg").Publish(ctx)
	assert.ErrorIs(t, err, ErrMissingChannel)
}

func TestRedisCache_Publish_MissingValue(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()
	err := With[string](cache).Channel("ch").Publish(ctx)
	assert.ErrorIs(t, err, ErrMissingValue)
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
	mock.ExpectSetNX("mykey", []byte("hello"), 0).SetVal(true)

	chain.Value("hello")
	ok, err := chain.SetIfNotExists(ctx)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())

	// key is existed return false
	mock.ExpectSetNX("mykey", []byte("world"), 0).SetVal(false)

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
	mock.ExpectSetNX("statusKey", []byte(StatusPending.String()), 0).SetVal(true)
	ok, err := chain.SetIfNotExists(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())

	chain.Key("statusKey").Value(StatusCompleted.String())
	mock.ExpectSetNX("statusKey", []byte(StatusCompleted.String()), 0).SetVal(false)
	ok, err = chain.SetIfNotExists(ctx)
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetIfNotExists_MissingKey(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()
	ok, err := With[string](cache).Value("x").SetIfNotExists(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.False(t, ok)
}

func TestSetIfNotExists_MissingValue(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()
	ok, err := With[string](cache).Key("k").SetIfNotExists(ctx)
	assert.ErrorIs(t, err, ErrMissingValue)
	assert.False(t, ok)
}

func TestRedisCache_Subscribe_MissingChannel(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()
	err := With[string](cache).Subscribe(ctx, func(string) {})
	assert.ErrorIs(t, err, ErrMissingChannel)
}

func TestRedisCache_New_NilConfig(t *testing.T) {
	c, err := New(nil)
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestRedisCache_Set_MissingKey(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	err := With[string](cache).Value("x").Set(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
}

func TestRedisCache_Set_MissingValue(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	err := With[string](cache).Key("k").Set(ctx)
	assert.ErrorIs(t, err, ErrMissingValue)
}

func TestRedisCache_Get_MissingKey(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	var zero string
	result, err := With[string](cache).Get(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.Equal(t, zero, result)
}

func TestRedisCache_GetMany(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	mock.ExpectMGet("k1", "k2", "k3").SetVal([]interface{}{"v1", nil, "v3"})

	vals, err := With[string](cache).Keys("k1", "k2", "k3").GetMany(ctx)
	require.NoError(t, err)
	require.Len(t, vals, 3)
	assert.Equal(t, "v1", vals[0])
	assert.Equal(t, "", vals[1]) // nil -> zero string
	assert.Equal(t, "v3", vals[2])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_GetMany_MissingKeys(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	vals, err := With[string](cache).GetMany(ctx)
	assert.ErrorIs(t, err, ErrMissingKeys)
	assert.Nil(t, vals)
}

func TestRedisCache_GetByPrefix_MissingPrefix(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	vals, err := With[string](cache).GetByPrefix(ctx)
	assert.ErrorIs(t, err, ErrMissingPrefix)
	assert.Nil(t, vals)
}

func TestRedisCache_SetMany(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	mock.ExpectSet("only", []byte("val"), 0).SetVal("OK")
	err := With[string](cache).Put("only", "val").SetMany(ctx)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_SetMany_EmptyBatch(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	err := With[string](cache).SetMany(ctx)
	assert.ErrorIs(t, err, ErrMissingPushOrBatch)
}

func TestRedisCache_Exists(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	mock.ExpectExists("exists_key").SetVal(1)
	ok, err := With[string](cache).Key("exists_key").Exists(ctx)
	require.NoError(t, err)
	assert.True(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectExists("missing_key").SetVal(0)
	ok, err = With[string](cache).Key("missing_key").Exists(ctx)
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisCache_Exists_MissingKey(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	ok, err := With[string](cache).Exists(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.False(t, ok)
}

func TestRedisCache_Delete_MissingKey(t *testing.T) {
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}
	ctx := context.Background()

	err := With[string](cache).Delete(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
}
