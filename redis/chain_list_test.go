package redis

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/go-redis/redismock/v9"
)

func TestChainList(t *testing.T) {
	ctx := context.Background()

	rdb, mock := redismock.NewClientMock()

	cache := &Cache{client: rdb,
		Config: &Config{
			TimeoutSec: 5,
		},
	}

	list := WithList[string](cache).Key("test:list")

	// Test Add (RPush)
	mock.ExpectRPush("test:list", "a", "b", "c").SetVal(3)

	err := list.Values([]string{
		"a", "b", "c",
	}).Add(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Size
	mock.ExpectLLen("test:list").SetVal(3)

	size, err := list.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), size)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Get index
	mock.ExpectLRange("test:list", int64(0), int64(0)).SetVal([]string{"a"})

	val, err := list.Get(ctx, 0)
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, "a", *val)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test GetRange
	mock.ExpectLRange("test:list", int64(0), int64(-1)).SetVal([]string{"a", "b", "c"})

	vals, err := list.GetRange(ctx)
	assert.NoError(t, err)
	assert.Len(t, vals, 3)
	assert.Equal(t, "a", *vals[0])
	assert.Equal(t, "b", *vals[1])
	assert.Equal(t, "c", *vals[2])
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test PopFront
	mock.ExpectLPop("test:list").SetVal("a")

	first, err := list.PopFront(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "a", *first)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Pop (RPop)
	mock.ExpectRPop("test:list").SetVal("c")

	last, err := list.Pop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "c", *last)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainList_WithStruct(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb,
		Config: &Config{
			TimeoutSec: 5,
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

	size, err := list.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), size)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Get ---
	mock.ExpectLRange("user:list", 0, 0).
		SetVal([]string{string(vals[0].([]byte))})

	val, err := list.Get(ctx, 0)
	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, users[0].ID, (*val).ID)
	assert.Equal(t, users[0].Name, (*val).Name)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- GetRange ---
	mock.ExpectLRange("user:list", int64(0), int64(-1)).
		SetVal([]string{string(vals[0].([]byte)), string(vals[1].([]byte))})

	valsOut, err := list.GetRange(ctx)
	assert.NoError(t, err)
	assert.Len(t, valsOut, 2)
	assert.Equal(t, users[0].ID, (*valsOut[0]).ID)
	assert.Equal(t, users[1].ID, (*valsOut[1]).ID)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- PopFront ---
	mock.ExpectLPop("user:list").
		SetVal(string(vals[0].([]byte)))

	first, err := list.PopFront(ctx)
	assert.NoError(t, err)
	assert.Equal(t, users[0].ID, (*first).ID)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Pop ---
	mock.ExpectRPop("user:list").
		SetVal(string(vals[1].([]byte)))
	last, err := list.Pop(ctx)
	assert.NoError(t, err)
	assert.Equal(t, users[1].ID, (*last).ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
