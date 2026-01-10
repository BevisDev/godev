package rabbitmq

import (
	"github.com/BevisDev/godev/consts"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildMessage_StringJSON(t *testing.T) {
	p := &Publisher{}

	ct, body, err := p.buildMessage(`{"a":1}`)

	require.NoError(t, err)
	require.Equal(t, consts.ApplicationJSON, ct)
	require.JSONEq(t, `{"a":1}`, string(body))
}

func TestBuildMessage_StringPlain(t *testing.T) {
	p := &Publisher{}

	ct, body, err := p.buildMessage("hello")

	require.NoError(t, err)
	require.Equal(t, consts.TextPlain, ct)
	require.Equal(t, "hello", string(body))
}

func TestBuildMessage_Number(t *testing.T) {
	p := &Publisher{}

	ct, body, err := p.buildMessage(123)

	require.NoError(t, err)
	require.Equal(t, consts.TextPlain, ct)
	require.Equal(t, "123", string(body))
}

func TestBuildMessage_Struct(t *testing.T) {
	p := &Publisher{}

	type payload struct {
		ID int `json:"id"`
	}

	ct, body, err := p.buildMessage(payload{ID: 1})

	require.NoError(t, err)
	require.Equal(t, consts.ApplicationJSON, ct)
	require.JSONEq(t, `{"id":1}`, string(body))
}

func TestBuildMessage_TooLarge(t *testing.T) {
	p := &Publisher{}

	large := make([]byte, maxMessageSize+1)

	_, _, err := p.buildMessage(large)

	require.Error(t, err)
	require.Contains(t, err.Error(), "message is too large")
}
