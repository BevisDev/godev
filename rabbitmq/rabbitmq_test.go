package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Defaults for integration tests against a local RabbitMQ broker.
// Override with env in the future if needed.
const (
	testRabbitHost     = "localhost"
	testRabbitPort     = 5672
	testRabbitUsername = "guest"
	testRabbitPassword = "guest"
	testRabbitVHost    = "/"
)

func testRabbitConfig() *Config {
	return &Config{
		Host:     testRabbitHost,
		Port:     testRabbitPort,
		Username: testRabbitUsername,
		Password: testRabbitPassword,
		VHost:    testRabbitVHost,
	}
}

// newTestMQ connects to RabbitMQ; skips the test if the broker is unavailable.
func newTestMQ(t *testing.T) *MQ {
	t.Helper()
	mq, err := New(context.Background(), testRabbitConfig())
	if err != nil {
		t.Skipf("skip when RabbitMQ is not available: %v", err)
	}
	t.Cleanup(func() { mq.Close() })
	return mq
}

func TestConnectMQ(t *testing.T) {
	mq := newTestMQ(t)

	require.NotNil(t, mq, "mq must not be nil")
	require.NotNil(t, mq.connection, "connection must not be nil")

	ch, err := mq.GetChannel()
	require.NoError(t, err, "should open channel")
	require.NotNil(t, ch, "channel must not be nil")

	_ = ch.Close()
}
