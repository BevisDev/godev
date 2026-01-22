package server

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_SetupCalled(t *testing.T) {
	setupCalled := false

	app := New(&Config{
		Port: "8080",
		Setup: func(r *gin.Engine) {
			setupCalled = true
		},
	})

	require.NotNil(t, app)
	assert.True(t, setupCalled)
}

func TestNew_NilConfig_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = New(nil)
	})
}

func TestHTTPApp_Stop_ShutdownCalled(t *testing.T) {
	shutdownCalled := false

	app := New(&Config{
		Port: "8080",
		Shutdown: func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := app.Stop(ctx)
	require.NoError(t, err)
	assert.True(t, shutdownCalled)
}

func TestHTTPApp_Run_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	app := New(&Config{Port: "8080"})

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := app.Run(ctx)
	require.NoError(t, err)
}

func TestHTTPApp_Stop_ShutdownTimeout(t *testing.T) {
	start := time.Now()

	app := New(&Config{
		Port:            "8080",
		ShutdownTimeout: 100 * time.Millisecond,
		Shutdown: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	})

	ctx := context.Background()

	err := app.Stop(ctx)
	require.NoError(t, err)

	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "shutdown timeout not respected")
}
