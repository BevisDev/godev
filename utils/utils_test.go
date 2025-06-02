package utils

import (
	"context"
	"errors"
	"github.com/BevisDev/godev/consts"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	Name string
	Age  int
}

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

func TestGetCurrentTimestamp(t *testing.T) {
	ts := GetCurrentTimestamp()
	assert.Greater(t, ts, int64(0))
}

func TestMaskLeft(t *testing.T) {
	tests := []struct {
		input    string
		size     int
		expected string
	}{
		{"abcdef", 3, "***def"},
		{"abcdef", 0, "******"},
		{"abcdef", 10, "******"},
		{"a", 1, "*"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MaskLeft(tt.input, tt.size)
			if result != tt.expected {
				t.Errorf("MaskLeft(%q, %d) = %q; want %q", tt.input, tt.size, result, tt.expected)
			}
		})
	}
}

func TestMaskRight(t *testing.T) {
	tests := []struct {
		input    string
		size     int
		expected string
	}{
		{"abcdef", 3, "abc***"},
		{"abcdef", 0, "******"},
		{"abcdef", 10, "******"},
		{"a", 1, "*"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MaskRight(tt.input, tt.size)
			if result != tt.expected {
				t.Errorf("MaskRight(%q, %d) = %q; want %q", tt.input, tt.size, result, tt.expected)
			}
		})
	}
}

func TestMaskCenter(t *testing.T) {
	tests := []struct {
		input    string
		size     int
		expected string
	}{
		{"abcdef", 2, "ab**ef"},
		{"abcdef", 3, "a***ef"},
		{"abcdef", 0, "******"},
		{"abcdef", 6, "******"},
		{"abcdef", 10, "******"},
		{"abc", 1, "a*c"},
		{"a", 1, "*"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MaskCenter(tt.input, tt.size)
			if result != tt.expected {
				t.Errorf("MaskCenter(%q, %d) = %q; want %q", tt.input, tt.size, result, tt.expected)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		sizeLocal  int
		sizeDomain int
		expected   string
	}{
		{
			name:       "Basic masking",
			email:      "john.doe@example.com",
			sizeLocal:  3,
			sizeDomain: 4,
			expected:   "john.***@example****",
		},
		{
			name:       "Mask full local",
			email:      "abc@example.com",
			sizeLocal:  10,
			sizeDomain: 4,
			expected:   "***@example****",
		},
		{
			name:       "Mask nothing",
			email:      "user@domain.com",
			sizeLocal:  0,
			sizeDomain: 0,
			expected:   "user@domain.com",
		},
		{
			name:       "Mask whole domain",
			email:      "someone@short.io",
			sizeLocal:  3,
			sizeDomain: 100,
			expected:   "some***@********",
		},
		{
			name:       "Invalid email format",
			email:      "invalid-email",
			sizeLocal:  3,
			sizeDomain: 3,
			expected:   "invalid-email",
		},
		{
			name:       "Short local and domain",
			email:      "a@b",
			sizeLocal:  1,
			sizeDomain: 1,
			expected:   "*@*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEmail(tt.email, tt.sizeLocal, tt.sizeDomain)
			if result != tt.expected {
				t.Errorf("MaskEmail(%q, %d, %d) = %q; want %q",
					tt.email, tt.sizeLocal, tt.sizeDomain, result, tt.expected)
			}
		})
	}
}

func TestIgnoreContentTypeLog(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"Image PNG", "image/png", true},
		{"Video MP4", "video/mp4", true},
		{"Audio MP3", "audio/mpeg", true},
		{"PDF", "application/pdf", true},
		{"Zip", "application/zip", true},
		{"VND Excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true},
		{"Protobuf", "application/x-protobuf", true},
		{"Binary", "application/octet-stream", true},
		{"Form-data", "multipart/form-data", true},
		{"JSON", "application/json", false},
		{"JSON with charset", "application/json; charset=utf-8", false},
		{"Plain text", "text/plain", false},
		{"Unknown type", "application/unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IgnoreContentTypeLog(tt.contentType); got != tt.want {
				t.Errorf("IgnoreContentTypeLog(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestParse_Success(t *testing.T) {
	obj := 123
	val, err := Parse[int](obj)

	assert.NoError(t, err)
	assert.Equal(t, 123, val)
}

func TestParse_Fail(t *testing.T) {
	obj := "abc"
	val, err := Parse[int](obj)

	assert.Error(t, err)
	assert.Equal(t, 0, val)
}

func TestParse_WithStruct(t *testing.T) {
	obj := User{Name: "Alice", Age: 30}

	val, err := Parse[User](obj)

	assert.NoError(t, err)
	assert.Equal(t, "Alice", val.Name)
	assert.Equal(t, 30, val.Age)
}

func TestParse_WithPointer(t *testing.T) {
	obj := &User{Name: "Bob", Age: 25}

	val, err := Parse[*User](obj)

	assert.NoError(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, "Bob", val.Name)
	assert.Equal(t, 25, val.Age)
}

func TestParse_Struct_CastFail(t *testing.T) {
	var obj interface{} = "not a User"
	val, err := Parse[User](obj)

	assert.Error(t, err)
	assert.Equal(t, User{}, val) // zero value
}

func TestParse_Pointer_CastFail(t *testing.T) {
	var obj interface{} = "not a *User"
	val, err := Parse[*User](obj)

	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestParseMap_Success(t *testing.T) {
	m := M{
		"user": User{Name: "Alice", Age: 30},
	}

	val, err := ParseMap[User]("user", m)

	assert.NoError(t, err)
	assert.Equal(t, "Alice", val.Name)
	assert.Equal(t, 30, val.Age)
}

func TestParseMap_MissingKey(t *testing.T) {
	m := M{}

	val, err := ParseMap[int]("notfound", m)

	assert.Error(t, err)
	assert.Equal(t, 0, val)
}

func TestParseMap_TypeMismatch(t *testing.T) {
	m := M{
		"age": "not-an-int",
	}

	val, err := ParseMap[int]("age", m)

	assert.Error(t, err)
	assert.Equal(t, 0, val)
}
