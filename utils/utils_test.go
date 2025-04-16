package utils

import (
	"context"
	"errors"
	"github.com/BevisDev/godev/consts"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetState_WhenCtxNil(t *testing.T) {
	state := GetState(nil)
	if state == "" {
		t.Error("Expected non-empty state")
	}
}

func TestGetState_WhenCtxHasNoState(t *testing.T) {
	ctx := context.Background()
	state := GetState(ctx)
	if state == "" {
		t.Error("Expected generated state")
	}
}

func TestGetState_WhenCtxHasState(t *testing.T) {
	expected := "fixed-state"
	ctx := context.WithValue(context.Background(), consts.State, expected)
	state := GetState(ctx)
	if state != expected {
		t.Errorf("GetState() = %q; want %q", state, expected)
	}
}

func TestCreateCtx_ShouldReturnContextWithState(t *testing.T) {
	ctx := CreateCtx()
	state := ctx.Value(consts.State)

	if state == nil || state == "" {
		t.Error("Expected state in context")
	}
}

func TestCreateCtxTimeout(t *testing.T) {
	ctx, cancel := CreateCtxTimeout(nil, 1)
	defer cancel()

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout context did not expire")
	}
}

func TestCreateCtxCancel(t *testing.T) {
	ctx, cancel := CreateCtxCancel(nil)
	cancel()

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Errorf("Expected context.Canceled, got %v", ctx.Err())
		}
	case <-time.After(1 * time.Second):
		t.Error("Cancel context did not close")
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	assert.True(t, ContainsIgnoreCase("Hello World", "hello"))
	assert.True(t, ContainsIgnoreCase("GoLang Is Fun", "IS"))
	assert.True(t, ContainsIgnoreCase("ABC", "abc"))
	assert.False(t, ContainsIgnoreCase("ABC", "xyz"))
	assert.False(t, ContainsIgnoreCase("hello", "world"))
}
