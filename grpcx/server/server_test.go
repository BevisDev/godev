package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestNew_SetupCalled(t *testing.T) {
	setupCalled := false

	app, err := New(&Config{
		Address: "127.0.0.1:0",
		Setup: func(s *grpc.Server) {
			setupCalled = s != nil
		},
	})

	require.NoError(t, err)
	require.NotNil(t, app)
	assert.True(t, setupCalled)
}

func TestNew_NilConfig_Error(t *testing.T) {
	app, err := New(nil)
	require.Error(t, err)
	assert.Nil(t, app)
}

func TestGRPCApp_StartAndStop(t *testing.T) {
	app, err := New(&Config{
		Address: "127.0.0.1:0",
	})
	require.NoError(t, err)

	err = app.Start()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestGRPCApp_Stop_ShutdownCalled(t *testing.T) {
	shutdownCalled := false

	app, err := New(&Config{
		Address: "127.0.0.1:0",
		Shutdown: func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		},
	})
	require.NoError(t, err)

	require.NoError(t, app.Start())
	err = app.Stop(context.Background())
	require.NoError(t, err)
	assert.True(t, shutdownCalled)
}

func TestGRPCApp_Run_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	app, err := New(&Config{
		Address: "127.0.0.1:0",
	})
	require.NoError(t, err)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err = app.Run(ctx)
	require.NoError(t, err)
}

func TestGRPCApp_Stop_ShutdownTimeout(t *testing.T) {
	start := time.Now()

	app, err := New(&Config{
		Address:         "127.0.0.1:0",
		ShutdownTimeout: 100 * time.Millisecond,
		Shutdown: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	})
	require.NoError(t, err)

	require.NoError(t, app.Start())
	err = app.Stop(context.Background())
	require.NoError(t, err)

	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "shutdown timeout not respected")
}
