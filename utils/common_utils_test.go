package utils

import (
	"context"
	"errors"
	"github.com/BevisDev/godev/constants"
	"regexp"
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
	ctx := context.WithValue(context.Background(), constants.State, expected)
	state := GetState(ctx)
	if state != expected {
		t.Errorf("GetState() = %q; want %q", state, expected)
	}
}

func TestCreateCtx_ShouldReturnContextWithState(t *testing.T) {
	ctx := CreateCtx()
	state := ctx.Value(constants.State)

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

func TestGenUUID(t *testing.T) {
	uuid := GenUUID()
	if uuid == "" {
		t.Errorf("GenUUID() = empty string")
	}
	r := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
	if !r.MatchString(uuid) {
		t.Errorf("GenUUID() = %q, not a valid UUID v4", uuid)
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"int", 123, "123"},
		{"int64", int64(999), "999"},
		{"float64", 3.14159, "3.14159"},
		{"float32", float32(1.618), "1.618"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"string", "hello", "hello"},
		{"[]byte", []byte("world"), "world"},
		{"nil", nil, ""},

		// pointer cases
		{"*int", func() any { v := 42; return &v }(), "42"},
		{"*int64", func() any { v := int64(888); return &v }(), "888"},
		{"*string", func() any { s := "pointer"; return &s }(), "pointer"},
		{"*bool", func() any { b := true; return &b }(), "true"},
		{"*float64", func() any { f := 3.3; return &f }(), "3.3"},
		{"nil *int", func() any { var p *int = nil; return p }(), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.input)
			if result != tt.expected {
				t.Errorf("ToString(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeToASCII(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cà phê sữa đá", "Ca phe sua da"},
		{"Thành phố Hồ Chí Minh", "Thanh pho Ho Chi Minh"},
		{"Điện Biên Phủ", "Dien Bien Phu"},
		{"Tôi yêu tiếng Việt", "Toi yeu tieng Viet"},
		{"Không dấu", "Khong dau"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeToASCII(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeToASCII(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cà-phê! sữa. đá?", "Caphe sua da"},
		{"Hello @#$%^&*()", "Hello "},
		{"Tên tôi là: Trần Văn *A*", "Ten toi la Tran Van A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CleanText(tt.input)
			if result != tt.expected {
				t.Errorf("CleanText(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveWhiteSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "HelloWorld"},
		{"  A B  C ", "ABC"},
		{"Không có khoảng trắng", "Khôngcókhoảngtrắng"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RemoveWhiteSpace(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveWhiteSpace(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{
			name:     "Short string",
			input:    "Hello",
			max:      10,
			expected: "Hello",
		},
		{
			name:     "Exact length",
			input:    "HelloWorld",
			max:      10,
			expected: "HelloWorld",
		},
		{
			name:     "Long string",
			input:    "Hello, this is a long message",
			max:      5,
			expected: "Hello",
		},
		{
			name:     "Empty string",
			input:    "",
			max:      10,
			expected: "",
		},
		{
			name:     "Zero max",
			input:    "Hello",
			max:      0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := TruncateText(tt.input, tt.max)
			if actual != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, actual)
			}
		})
	}
}
