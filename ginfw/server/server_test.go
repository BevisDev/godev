package server

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNew_SetupCalled(t *testing.T) {
	setupCalled := false

	app := New(&Config{
		Port: "8080",
		Setup: func(r *gin.Engine) {
			setupCalled = true
		},
	})

	if app == nil {
		t.Fatal("expected HTTPApp to be created")
	}

	if !setupCalled {
		t.Fatal("expected Setup to be called")
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !shutdownCalled {
		t.Fatal("expected Shutdown hook to be called")
	}
}

func TestHTTPApp_Run_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	app := New(&Config{
		Port: "8080",
	})

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := app.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPApp_Stop_ShutdownTimeout(t *testing.T) {
	start := time.Now()

	app := New(&Config{
		Port:            "8080",
		ShutdownTimeout: 100 * time.Millisecond,
		Shutdown: func(ctx context.Context) error {
			// Wait for context to be cancelled (by timeout)
			<-ctx.Done()
			return ctx.Err()
		},
	})

	// Use background context (not cancelled) so that WithTimeout can create a proper timeout
	ctx := context.Background()

	err := app.Stop(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Fatalf("shutdown timeout not respected: elapsed %v, expected at least 100ms", elapsed)
	}
}
