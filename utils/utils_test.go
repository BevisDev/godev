package utils

import (
	"context"
	"testing"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

type User struct {
	Name string
	Age  int
}

func TestNewCtx_ShouldReturnContextWithRID(t *testing.T) {
	ctx := NewCtx()
	state := ctx.Value(consts.RID)

	if state == nil || state == "" {
		t.Error("Expected state in context")
	}
}

func TestGetRID_WhenCtxHasRID(t *testing.T) {
	expected := "fixed-rid"
	ctx := context.WithValue(context.Background(), consts.RID, expected)
	rid := GetRID(ctx)
	if rid != expected {
		t.Errorf("GetRID() = %q; want %q", rid, expected)
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	assert.True(t, ContainsIgnoreCase("Hello World", "hello"))
	assert.True(t, ContainsIgnoreCase("GoLang Is Fun", "IS"))
	assert.True(t, ContainsIgnoreCase("ABC", "abc"))
	assert.False(t, ContainsIgnoreCase("ABC", "xyz"))
	assert.False(t, ContainsIgnoreCase("hello", "world"))
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
			if got := SkipContentType(tt.contentType); got != tt.want {
				t.Errorf("skipContentType(%q) = %v, want %v", tt.contentType, got, tt.want)
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
	m := types.Object{
		"user": User{Name: "Alice", Age: 30},
	}

	val, err := ParseValueMap[User]("user", m)

	assert.NoError(t, err)
	assert.Equal(t, "Alice", val.Name)
	assert.Equal(t, 30, val.Age)
}

func TestParseValueMap_MissingKey(t *testing.T) {
	m := types.Object{}

	val, err := ParseValueMap[int]("notfound", m)

	assert.Error(t, err)
	assert.Equal(t, 0, val)
}

func TestParseValueMap_TypeMismatch(t *testing.T) {
	m := types.Object{
		"age": "not-an-int",
	}

	val, err := ParseValueMap[int]("age", m)

	assert.Error(t, err)
	assert.Equal(t, 0, val)
}

func TestIsContains(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		value    int
		expected bool
	}{
		{"value exists", []int{1, 2, 3}, 2, true},
		{"value not exists", []int{1, 2, 3}, 5, false},
		{"empty array", []int{}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsContains(tt.arr, tt.value)
			if result != tt.expected {
				t.Errorf("IsContains(%v, %v) = %v; want %v", tt.arr, tt.value, result, tt.expected)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	slice := []int{10, 20, 30, 40}

	assert.Equal(t, 2, IndexOf(slice, 30))
	assert.Equal(t, -1, IndexOf(slice, 100))
}

func TestPercent(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected float64
	}{
		{"percent of int", 65, 0.65},
		{"percent of int8", int8(25), 0.25},
		{"percent of int16", int16(100), 1.0},
		{"percent of int32", int32(3), 0.03},
		{"percent of int64", int64(0), 0.0},
		{"percent negative", -20, -0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual float64

			switch v := tt.input.(type) {
			case int:
				actual = Percent(v)
			case int8:
				actual = Percent(v)
			case int16:
				actual = Percent(v)
			case int32:
				actual = Percent(v)
			case int64:
				actual = Percent(v)
			default:
				t.Fatalf("unsupported type: %T", v)
			}

			assert.InDelta(t, tt.expected, actual, 0.00001)
		})
	}
}

func TestMillion(t *testing.T) {
	type testCase struct {
		name     string
		input    int64
		expected int64
	}

	tests := []testCase{
		{"zero", 0, 0},
		{"one", 1, 1_000_000},
		{"five", 5, 5_000_000},
		{"hundred", 100, 100_000_000},
		{"negative", -3, -3_000_000},
		{"large", 1_234_567, 1_234_567_000_000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Million(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRoundDownToMul(t *testing.T) {
	type testCase[T ~int | ~int8 | ~int16 | ~int32 | ~int64] struct {
		name     string
		input    T
		mul      T
		expected T
	}

	tests := []testCase[int]{
		{"round 0", 0, 5, 0},
		{"round exact", 5, 5, 5},
		{"round under", 3, 5, 0},
		{"round 42 → 40", 42, 5, 40},
		{"round 45 → 45", 45, 5, 45},
		{"round 47 → 45", 47, 5, 45},
		{"round 99 → 95", 99, 5, 95},
		{"round 100 → 100", 100, 5, 100},
		{"round large million", 42_000_000, 5_000_000, 40_000_000},
		{"round large exact", 45_000_000, 5_000_000, 45_000_000},
		{"round negative -7 → -10", -7, 5, -10},
		{"round negative -4 → -5", -4, 5, -5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := RoundDownToMul(tc.input, tc.mul)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRoundUpToMul(t *testing.T) {
	type testCase[T constraints.Integer] struct {
		n        T
		mul      T
		expected T
	}

	tests := []testCase[int]{
		{47, 5, 50},
		{40, 10, 40},
		{41, 10, 50},
		{0, 5, 0},
		{5, 5, 5},
		{42_000_000, 5_000_000, 45_000_000},
		{47, 0, 47},
		{47, -5, 47},
		{-1, 5, 0},
		{-4, 5, 0},
		{-5, 5, -5},
		{-6, 5, -5},
		{-7, 5, -5},
		{-11, 5, -10},
		{-13, 5, -10},
		{-15, 5, -15},
		{-17, 5, -15},
	}

	for _, tc := range tests {
		result := RoundUpToMul(tc.n, tc.mul)
		if result != tc.expected {
			t.Errorf("RoundUpToMul(%d, %d) = %d; want %d", tc.n, tc.mul, result, tc.expected)
		}
	}
}

func TestPtrTo(t *testing.T) {
	const hello = "hello"

	t.Run("const", func(t *testing.T) {
		p := GetPointer(hello)
		assert.NotNil(t, p)
		assert.Equal(t, "hello", *p)
	})

	t.Run("string", func(t *testing.T) {
		s := "hello"
		p := GetPointer(s)
		assert.NotNil(t, p)
		assert.Equal(t, "hello", *p)
	})

	t.Run("int", func(t *testing.T) {
		n := 42
		p := GetPointer(n)
		assert.NotNil(t, p)
		assert.Equal(t, 42, *p)
	})

	t.Run("struct", func(t *testing.T) {
		type User struct {
			ID int
		}
		u := User{ID: 1}
		p := GetPointer(u)
		assert.NotNil(t, p)
		assert.Equal(t, 1, p.ID)
	})
}

func TestToBytes(t *testing.T) {
	t.Run("nil value returns error", func(t *testing.T) {
		b, err := ToBytes(nil)
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Contains(t, err.Error(), "value is nil")
	})

	t.Run("nil pointer returns error", func(t *testing.T) {
		var p *int
		b, err := ToBytes(p)
		assert.Error(t, err)
		assert.Nil(t, b)
		assert.Contains(t, err.Error(), "value is nil")
	})

	t.Run("nil slice returns error", func(t *testing.T) {
		var s []byte
		b, err := ToBytes(s)
		assert.Error(t, err)
		assert.Nil(t, b)
	})

	t.Run("[]byte returns as-is", func(t *testing.T) {
		raw := []byte("hello")
		b, err := ToBytes(raw)
		assert.NoError(t, err)
		assert.Equal(t, []byte("hello"), b)
	})

	t.Run("empty []byte returns as-is", func(t *testing.T) {
		raw := []byte{}
		b, err := ToBytes(raw)
		assert.NoError(t, err)
		assert.Equal(t, []byte{}, b)
	})

	t.Run("string to bytes", func(t *testing.T) {
		b, err := ToBytes("hello")
		assert.NoError(t, err)
		assert.Equal(t, []byte("hello"), b)
	})

	t.Run("empty string to bytes", func(t *testing.T) {
		b, err := ToBytes("")
		assert.NoError(t, err)
		assert.Equal(t, []byte(""), b)
	})

	t.Run("bool to bytes", func(t *testing.T) {
		bt, err := ToBytes(true)
		assert.NoError(t, err)
		assert.Equal(t, []byte("true"), bt)

		bf, err := ToBytes(false)
		assert.NoError(t, err)
		assert.Equal(t, []byte("false"), bf)
	})

	t.Run("int types to bytes", func(t *testing.T) {
		cases := []struct {
			value    any
			expected string
		}{
			{42, "42"},
			{int8(-1), "-1"},
			{int16(100), "100"},
			{int32(0), "0"},
			{int64(999), "999"},
			{uint(1), "1"},
			{uint8(255), "255"},
			{uint16(65535), "65535"},
			{uint32(0), "0"},
			{uint64(1), "1"},
		}
		for _, c := range cases {
			b, err := ToBytes(c.value)
			assert.NoError(t, err)
			assert.Equal(t, []byte(c.expected), b, "value %v", c.value)
		}
	})

	t.Run("float types to bytes", func(t *testing.T) {
		b32, err := ToBytes(float32(3.14))
		assert.NoError(t, err)
		assert.Equal(t, []byte("3.14"), b32)

		b64, err := ToBytes(3.14)
		assert.NoError(t, err)
		assert.Equal(t, []byte("3.14"), b64)
	})

	t.Run("struct to JSON bytes", func(t *testing.T) {
		u := User{Name: "Alice", Age: 30}
		b, err := ToBytes(u)
		assert.NoError(t, err)
		assert.Equal(t, []byte(`{"Name":"Alice","Age":30}`), b)
	})

	t.Run("map to JSON bytes", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		b, err := ToBytes(m)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"a":1,"b":2}`, string(b))
	})

	t.Run("pointer to struct to JSON bytes", func(t *testing.T) {
		u := &User{Name: "Bob", Age: 25}
		b, err := ToBytes(u)
		assert.NoError(t, err)
		assert.Equal(t, []byte(`{"Name":"Bob","Age":25}`), b)
	})
}
