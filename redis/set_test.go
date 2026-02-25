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

	// --- Test Add (values stored as []byte via utils.ToBytes)
	mock.ExpectSAdd("test:set", []byte("a"), []byte("b"), []byte("c")).SetVal(3)
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
	mock.ExpectSIsMember("test:set", []byte("a")).SetVal(true)
	ok, err := set.Key("test:set").Contains(ctx, "a")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Remove
	mock.ExpectSRem("test:set", []byte("a")).SetVal(1)
	err = set.Key("test:set").Values("a").Remove(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test GetAll (after Remove("a"), set has "b", "c")
	mock.ExpectSMembers("test:set").SetVal([]string{"b", "c"})
	vals, err := set.Key("test:set").GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, vals, 2)
	assert.Equal(t, "b", vals[0])
	assert.Equal(t, "c", vals[1])
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

	// Chuẩn bị giá trị dạng []byte (utils.ToBytes Marshal JSON)
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
	assert.Equal(t, users[0].Name, valsOut[0].Name)
	assert.Equal(t, users[1].Name, valsOut[1].Name)
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

	// utils.ToBytes marshals Status (custom type) as JSON, so stored as "pending", "approved" in JSON
	pendingJSON, _ := json.Marshal(string(StatusPending))
	approvedJSON, _ := json.Marshal(string(StatusApproved))

	// --- Test Add
	mock.ExpectSAdd("status:set", pendingJSON, approvedJSON).SetVal(2)
	err := set.Values(statuses).Add(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test Contains
	mock.ExpectSIsMember("status:set", pendingJSON).SetVal(true)
	ok, err := set.Key("status:set").Contains(ctx, StatusPending)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())

	// --- Test GetAll (Redis returns strings; JSON bytes become string form)
	mock.ExpectSMembers("status:set").SetVal([]string{string(pendingJSON), string(approvedJSON)})
	vals, err := set.Key("status:set").GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, vals, 2)
	assert.Equal(t, StatusPending, vals[0])
	assert.Equal(t, StatusApproved, vals[1])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainSet_ErrorCases(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{
		client: rdb,
		cf:     &Config{Timeout: 5 * time.Second},
	}

	t.Run("missing key", func(t *testing.T) {
		set := WithSet[string](cache)
		err := set.Values("a").Add(ctx)
		assert.ErrorIs(t, err, ErrMissingKey)
	})

	t.Run("missing values", func(t *testing.T) {
		set := WithSet[string](cache)
		err := set.Key("test:set").Add(ctx)
		assert.ErrorIs(t, err, ErrMissingValues)
	})
}

func TestChainSet_Contains_False(t *testing.T) {
	ctx := context.Background()
	rdb, mock := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	mock.ExpectSIsMember("s", []byte("missing")).SetVal(false)
	ok, err := WithSet[string](cache).Key("s").Contains(ctx, "missing")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChainSet_Contains_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	ok, err := WithSet[string](cache).Contains(ctx, "x")
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.False(t, ok)
}

func TestChainSet_Remove_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithSet[string](cache).Values("a").Remove(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
}

func TestChainSet_Remove_MissingValues(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithSet[string](cache).Key("s").Remove(ctx)
	assert.ErrorIs(t, err, ErrMissingValues)
}

func TestChainSet_GetAll_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	vals, err := WithSet[string](cache).GetAll(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.Nil(t, vals)
}

func TestChainSet_Size_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	size, err := WithSet[string](cache).Size(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
	assert.Equal(t, int64(0), size)
}

func TestChainSet_Delete_MissingKey(t *testing.T) {
	ctx := context.Background()
	rdb, _ := redismock.NewClientMock()
	cache := &Cache{client: rdb, cf: &Config{Timeout: 5 * time.Second}}

	err := WithSet[string](cache).Delete(ctx)
	assert.ErrorIs(t, err, ErrMissingKey)
}
