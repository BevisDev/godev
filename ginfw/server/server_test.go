package server

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRun_SetupCalled(t *testing.T) {
	setupCalled := false

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, Options{
		Port: "8080",
		Setup: func(r *gin.Engine) {
			setupCalled = true
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !setupCalled {
		t.Fatal("expected Setup to be called")
	}
}

func TestRun_ShutdownCalled(t *testing.T) {
	shutdownCalled := false

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, Options{
		Port: "8080",
		Shutdown: func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !shutdownCalled {
		t.Fatal("expected Shutdown hook to be called")
	}
}

func TestRun_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		cancel()
	}()

	err := Run(ctx, Options{
		Port: "8080",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_ShutdownTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()

	err := Run(ctx, Options{
		Port:            "8080",
		ShutdownTimeout: 100 * time.Millisecond,
		Shutdown: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if time.Since(start) < 100*time.Millisecond {
		t.Fatal("shutdown timeout not respected")
	}
}
