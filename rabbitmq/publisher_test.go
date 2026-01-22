package rabbitmq

import (
	"testing"

	"github.com/BevisDev/godev/consts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMessage_StringJSON(t *testing.T) {
	p := &Publisher{}
	ct, body, err := p.buildMessage(`{"a":1}`)
	require.NoError(t, err)
	assert.Equal(t, consts.ApplicationJSON, ct)
	require.JSONEq(t, `{"a":1}`, string(body))
}

func TestBuildMessage_StringPlain(t *testing.T) {
	p := &Publisher{}
	ct, body, err := p.buildMessage("hello")
	require.NoError(t, err)
	assert.Equal(t, consts.TextPlain, ct)
	assert.Equal(t, "hello", string(body))
}

func TestBuildMessage_Number(t *testing.T) {
	p := &Publisher{}
	ct, body, err := p.buildMessage(123)
	require.NoError(t, err)
	assert.Equal(t, consts.TextPlain, ct)
	assert.Equal(t, "123", string(body))
}

func TestBuildMessage_Struct(t *testing.T) {
	p := &Publisher{}
	type payload struct {
		ID int `json:"id"`
	}
	ct, body, err := p.buildMessage(payload{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, consts.ApplicationJSON, ct)
	require.JSONEq(t, `{"id":1}`, string(body))
}

func TestBuildMessage_Map(t *testing.T) {
	p := &Publisher{}
	ct, body, err := p.buildMessage(map[string]int{"a": 1, "b": 2})
	require.NoError(t, err)
	assert.Equal(t, consts.ApplicationJSON, ct)
	require.JSONEq(t, `{"a":1,"b":2}`, string(body))
}

func TestBuildMessage_Slice(t *testing.T) {
	p := &Publisher{}
	ct, body, err := p.buildMessage([]string{"x", "y", "z"})
	require.NoError(t, err)
	assert.Equal(t, consts.ApplicationJSON, ct)
	require.JSONEq(t, `["x","y","z"]`, string(body))
}

func TestBuildMessage_TooLarge(t *testing.T) {
	p := &Publisher{}
	large := make([]byte, maxMessageSize+1)
	_, _, err := p.buildMessage(large)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message is too large")
}

func TestRabbitMQ_New_NilConfig(t *testing.T) {
	mq, err := New(nil)
	require.Error(t, err)
	assert.Nil(t, mq)
	assert.Contains(t, err.Error(), "config is nil")
}
