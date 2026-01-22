package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestChainSet_StringValue(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf:     &Config{Timeout: 5 * time.Second},
	}

	set := WithSet[string](cache).Key("test:set")

	// --- Test Add
	mock.ExpectSAdd("test:set", "a", "b", "c").SetVal(3)
	err := set.Values([]string{"a", "b", "c"}).Add(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Size
	mock.ExpectSCard("test:set").SetVal(3)
	size, err := set.Key("test:set").Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), size)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Contains true
	mock.ExpectSIsMember("test:set", "a").SetVal(true)
	ok, err := set.Key("test:set").Contains(ctx, "a")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Remove
	mock.ExpectSRem("test:set", "a").SetVal(1)
	err = set.Key("test:set").Values("a").Remove(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test GetAll
	mock.ExpectSMembers("test:set").SetVal([]string{"a", "b", "c"})
	vals, err := set.Key("test:set").GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, vals, 3)
	assert.Equal(t, "a", *vals[0])
	assert.Equal(t, "b", *vals[1])
	assert.Equal(t, "c", *vals[2])
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Delete
	mock.ExpectDel("test:set").SetVal(1)
	err = set.Key("test:set").Delete(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainSet_StructValue(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf:     &Config{Timeout: 5 * time.Second},
	}

	set := WithSet[User](cache).Key("user:set")

	users := []*User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}

	// Chuẩn bị giá trị dạng []byte (vì convertValue sẽ Marshal JSON)
	vals := make([]interface{}, len(users))
	for i, u := range users {
		b, _ := json.Marshal(u)
		vals[i] = b
	}

	// --- Test Add
	mock.ExpectSAdd("user:set", vals...).SetVal(2)
	err := set.Values(users).Add(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Size
	mock.ExpectSCard("user:set").SetVal(2)
	size, err := set.Key("user:set").Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), size)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Contains
	mock.ExpectSIsMember("user:set", vals[0]).SetVal(true)
	ok, err := set.Key("user:set").Contains(ctx, users[0])
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test GetAll
	mock.ExpectSMembers("user:set").SetVal([]string{
		string(vals[0].([]byte)),
		string(vals[1].([]byte)),
	})
	valsOut, err := set.Key("user:set").GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, valsOut, 2)
	assert.Equal(t, users[0].Name, (*valsOut[0]).Name)
	assert.Equal(t, users[1].Name, (*valsOut[1]).Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainSet_WithEnum(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf:     &Config{Timeout: 5 * time.Second},
	}

	set := WithSet[Status](cache).Key("status:set")

	statuses := []Status{StatusPending, StatusApproved}

	// --- Test Add
	mock.ExpectSAdd("status:set", string(StatusPending), string(StatusApproved)).SetVal(2)
	err := set.Values(statuses).Add(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Contains
	mock.ExpectSIsMember("status:set", string(StatusPending)).SetVal(true)
	ok, err := set.Key("status:set").Contains(ctx, StatusPending)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test GetAll
	mock.ExpectSMembers("status:set").SetVal([]string{"pending", "approved"})
	vals, err := set.Key("status:set").GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, vals, 2)
	assert.Equal(t, StatusPending, *vals[0])
	assert.Equal(t, StatusApproved, *vals[1])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainSet_ErrorCases(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf:     &Config{Timeout: 5 * time.Second},
	}

	set := WithSet[string](cache)

	t.Run("missing key", func(t *testing.T) {
		err := set.Values("a").Add(ctx)
		assert.ErrorIs(t, err, ErrMissingKey)
	})

	t.Run("missing values", func(t *testing.T) {
		err := set.Key("test:set").Add(ctx)
		assert.ErrorIs(t, err, ErrMissingValues)
	})
}
