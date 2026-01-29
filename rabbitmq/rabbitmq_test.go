package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewMQTest() (*MQ, error) {
	mq, err := New(&Config{
		Host:     "localhost",
		Port:     5672,
		Username: "admin",
		Password: "pass123",
		VHost:    "de-dev",
	})
	return mq, err
}

func TestConnectMQ(t *testing.T) {
	mq, err := NewMQTest()

	require.NoError(t, err, "should connect to rabbitmq")
	require.NotNil(t, mq, "mq must not be nil")
	require.NotNil(t, mq.connection, "connection must not be nil")

	// get channel to check connection alive
	ch, err := mq.GetChannel()
	require.NoError(t, err, "should open channel")
	require.NotNil(t, ch, "channel must not be nil")

	_ = ch.Close()
	mq.Close()
}
