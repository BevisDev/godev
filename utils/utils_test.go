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

func mustParse(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return t
}

func TestCalcAgeAt(t *testing.T) {
	tests := []struct {
		dob      string
		now      string
		expected int
		name     string
	}{
		{"2000-04-20", "2025-04-21", 25, "Birthday passed this year"},
		{"2000-05-10", "2025-04-21", 24, "Birthday not yet this year"},
		{"2000-04-21", "2025-04-21", 25, "Birthday is today"},
		{"2025-04-21", "2025-04-21", 0, "Born today"},
		{"2026-01-01", "2025-04-21", -1, "Future date"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dob := mustParse(tc.dob)
			now := mustParse(tc.now)
			age := CalcAgeAt(dob, now)
			if age != tc.expected {
				t.Errorf("Expected age %d, got %d", tc.expected, age)
			}
		})
	}
}
