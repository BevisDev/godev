package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainList(t *testing.T) {
	ctx := context.Background()

	rdb, mock := redismock.NewClientMock()

	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
		},
	}

	list := WithList[string](cache).Key("test:list")

	// Test Add (RPush) - values stored as []byte via utils.ToBytes
	mock.ExpectRPush("test:list", []byte("a"), []byte("b"), []byte("c")).SetVal(3)

	err := list.Values([]string{
		"a", "b", "c",
	}).Add(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Size
	mock.ExpectLLen("test:list").SetVal(3)

	size, err := list.Key("test:list").Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), size)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Get index
	mock.ExpectLRange("test:list", int64(0), int64(0)).SetVal([]string{"a"})

	val, err := list.Key("test:list").Get(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, "a", val)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test GetRange
	mock.ExpectLRange("test:list", int64(0), int64(-1)).SetVal([]string{"a", "b", "c"})

	vals, err := list.Key("test:list").Start(0).GetRange(ctx)
	assert.NoError(t, err)
	assert.Len(t, vals, 3)
	assert.Equal(t, "a", vals[0])
	assert.Equal(t, "b", vals[1])
	assert.Equal(t, "c", vals[2])
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test PopFront
	mock.ExpectLPop("test:list").SetVal("a")

	first, err := list.Key("test:list").PopFront(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "a", first)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Pop (RPop)
	mock.ExpectRPop("test:list").SetVal("c")

	last, err := list.Key("test:list").Pop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "c", last)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_WithStruct(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf: &Config{
			Timeout: 5 * time.Second,
		},
	}

	list := WithList[User](cache).Key("user:list")

	// Sample users
	users := []*User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}

	// Convert struct -> JSON string for mock
	vals := make([]interface{}, len(users))
	for i, u := range users {
		b, _ := json.Marshal(u)
		vals[i] = b
	}

	// Test Add (RPush)
	mock.ExpectRPush("user:list", vals...).SetVal(int64(len(users)))

	err := list.Values(users).Add(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Size ---
	mock.ExpectLLen("user:list").SetVal(int64(len(users)))

	size, err := list.Key("user:list").Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), size)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Get ---
	mock.ExpectLRange("user:list", 0, 0).
		SetVal([]string{string(vals[0].([]byte))})

	val, err := list.Key("user:list").Get(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, users[0].ID, val.ID)
	assert.Equal(t, users[0].Name, val.Name)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- GetRange ---
	mock.ExpectLRange("user:list", int64(0), int64(-1)).
		SetVal([]string{string(vals[0].([]byte)), string(vals[1].([]byte))})

	valsOut, err := list.Key("user:list").Start(0).GetRange(ctx)
	assert.NoError(t, err)
	assert.Len(t, valsOut, 2)
	assert.Equal(t, users[0].ID, valsOut[0].ID)
	assert.Equal(t, users[1].ID, valsOut[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- PopFront ---
	mock.ExpectLPop("user:list").
		SetVal(string(vals[0].([]byte)))

	first, err := list.Key("user:list").PopFront(ctx)
	assert.NoError(t, err)
	assert.Equal(t, users[0].ID, first.ID)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Pop ---
	mock.ExpectRPop("user:list").
		SetVal(string(vals[1].([]byte)))
	last, err := list.Key("user:list").Pop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, users[1].ID, last.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_AddFirst(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	mock.ExpectLPush("head:list", []byte("x"), []byte("y")).SetVal(2)
	err := WithList[string](cache).Key("head:list").Values([]string{"x", "y"}).AddFirst(ctx)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectLPop("head:list").SetVal("x")
	first, err := WithList[string](cache).Key("head:list").PopFront(ctx)
	require.NoError(t, err)
	assert.Equal(t, "x", first)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_Add_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithList[string](cache).Values([]string{"a"}).Add(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
}

func TestChainList_Add_MissingValues(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithList[string](cache).Key("k").Add(ctx)
	assert.ErrorIs(t, err, ErrMissingValues)
}

func TestChainList_AddFirst_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithList[string](cache).Values([]string{"a"}).AddFirst(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
}

func TestChainList_AddFirst_MissingValues(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithList[string](cache).Key("k").AddFirst(ctx)
	assert.ErrorIs(t, err, ErrMissingValues)
}

func TestChainList_Get_OutOfRange(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	mock.ExpectLRange("short:list", int64(10), int64(10)).SetVal([]string{})
	val, err := WithList[string](cache).Key("short:list").Get(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, "", val)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_PopFront_EmptyList(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	mock.ExpectLPop("empty:list").SetErr(redis.Nil)
	val, err := WithList[string](cache).Key("empty:list").PopFront(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", val)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_Pop_EmptyList(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	mock.ExpectRPop("empty:list").SetErr(redis.Nil)
	val, err := WithList[string](cache).Key("empty:list").Pop(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", val)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_Get_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	var zero string
	val, err := WithList[string](cache).Get(ctx, 0)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.Equal(t, zero, val)
}

func TestChainList_Delete(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	mock.ExpectDel("list:key").SetVal(1)
	err := WithList[string](cache).Key("list:key").Delete(ctx)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_Delete_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithList[string](cache).Delete(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
}

func TestChainList_Size_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	size, err := WithList[string](cache).Size(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.Equal(t, int64(0), size)
}
