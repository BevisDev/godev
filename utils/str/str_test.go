package str

import (
	"testing"
	"time"

	"github.com/BevisDev/godev/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func decimalPtr(f float64) *decimal.Decimal {
	d := decimal.NewFromFloat(f)
	return &d
}

type BadJSON struct {
	Ch chan int
}

func intPtrPtr(i int) any {
	p := &i
	return &p
}

func TestToString(t *testing.T) {
	now := time.Date(2024, 5, 1, 12, 30, 0, 0, time.UTC)
	dec := decimal.NewFromFloat(12.345)
	decPtr := decimalPtr(9.99)

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	user := User{"Bob", 30}

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		// primitives
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
		{"**int", intPtrPtr(77), "77"},
		{"*int64", func() any { v := int64(888); return &v }(), "888"},
		{"*string", func() any { s := "pointer"; return &s }(), "pointer"},
		{"*bool", func() any { b := true; return &b }(), "true"},
		{"*float64", func() any { f := 3.3; return &f }(), "3.3"},
		{"nil *int", func() any { var p *int = nil; return p }(), ""},

		// time cases
		{"time.Time", now, now.Format(time.RFC3339)},
		{"*time.Time", &now, now.Format(time.RFC3339)},

		// slice
		{"[]int", []int{1, 2, 3}, `[1,2,3]`},
		{"[]string", []string{"a", "b"}, `["a","b"]`},
		{"[]struct", []User{user}, `[{"name":"Bob","age":30}]`},
		{"empty slice", []int{}, `[]`},

		// array
		{"array", [3]int{1, 2, 3}, `[1,2,3]`},

		// map
		{"map[string]int", map[string]int{"a": 1}, `{"a":1}`},
		{"map empty", map[string]string{}, `{}`},

		// decimal.Decimal
		{"decimal.Decimal", dec, dec.String()},
		{"*decimal.Decimal", decPtr, decPtr.String()},

		// generic struct (JSON)
		{"struct as JSON", user, `{"name":"Bob","age":30}`},

		// fallback (marshal fail)
		{"bad json struct", BadJSON{}, "{Ch:<nil>}"},
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

func TestToInt(t *testing.T) {
	type testCase[T types.SignedInteger] struct {
		input    string
		expected T
	}

	tests := []testCase[int]{
		{"42", 42},
		{"0", 0},
		{"-123", -123},
		{"abc", 0}, // invalid string → fallback 0
		{"", 0},    // empty string → fallback 0
	}

	for _, tt := range tests {
		result := ToInt[int](tt.input)
		assert.Equal(t, tt.expected, result, "ToInt[int](%q)", tt.input)
	}
}

func TestToInt_Int8(t *testing.T) {
	type testCase struct {
		input    string
		expected int8
	}
	tests := []testCase{
		{"127", 127},   // max int8
		{"-128", -128}, // min int8
		{"0", 0},
		{"42", 42},
		{"128", 0},  // overflow
		{"-129", 0}, // underflow
		{"abc", 0},  // invalid
		{"", 0},     // empty string
	}

	for _, tt := range tests {
		result := ToInt[int8](tt.input)
		assert.Equal(t, tt.expected, result, "ToInt[int8](%q)", tt.input)
	}
}

func TestToInt_Int16(t *testing.T) {
	type testCase struct {
		input    string
		expected int16
	}
	tests := []testCase{
		{"123", 123},
		{"-32768", -32768},
		{"32767", 32767},
		{"32768", 0},  // overflow
		{"-32769", 0}, // underflow
		{"abc", 0},    // invalid
	}

	for _, tt := range tests {
		result := ToInt[int16](tt.input)
		assert.Equal(t, tt.expected, result, "ToInt[int16](%q)", tt.input)
	}
}

func TestToInt_Int32(t *testing.T) {
	type testCase struct {
		input    string
		expected int32
	}
	tests := []testCase{
		{"123456", 123456},
		{"2147483647", 2147483647},   // max int32
		{"-2147483648", -2147483648}, // min int32
		{"2147483648", 0},            // overflow
		{"abc", 0},                   // invalid
	}

	for _, tt := range tests {
		result := ToInt[int32](tt.input)
		assert.Equal(t, tt.expected, result, "ToInt[int32](%q)", tt.input)
	}
}

func TestToInt_Int32_Overflow(t *testing.T) {
	result := ToInt[int32]("9999999999")
	assert.Equal(t, int32(0), result, "should fallback to 0 due to overflow")
}

func TestToInt_Int64(t *testing.T) {
	type testCase struct {
		input    string
		expected int64
	}
	tests := []testCase{
		{"9223372036854775807", 9223372036854775807},   // max int64
		{"-9223372036854775808", -9223372036854775808}, // min int64
		{"9223372036854775808", 0},                     // overflow
		{"abc", 0},                                     // invalid
	}

	for _, tt := range tests {
		result := ToInt[int64](tt.input)
		assert.Equal(t, tt.expected, result, "ToInt[int64](%q)", tt.input)
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"0", 0.0},
		{"-2.718", -2.718},
		{"abc", 0.0},   // invalid → fallback 0
		{"", 0.0},      // empty → fallback 0
		{"1e10", 1e10}, // scientific notation
	}

	for _, tt := range tests {
		result := ToFloat[float64](tt.input)
		assert.InDelta(t, tt.expected, result, 0.0001, "ToFloat[float64](%q)", tt.input)
	}
}

func TestToFloat32(t *testing.T) {
	tests := []struct {
		input    string
		expected float32
	}{
		{"3.14", 3.14},
		{"0", 0.0},
		{"-2.5", -2.5},
		{"abc", 0.0},
		{"1.2e3", 1200.0},
	}

	for _, tt := range tests {
		result := ToFloat[float32](tt.input)
		assert.InDelta(t, tt.expected, result, 0.001, "ToFloat[float32](%q)", tt.input)
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid true lowercase", "true", true},
		{"Valid true uppercase", "TRUE", true},
		{"Valid true mixedcase", "TrUe", false},
		{"Valid false lowercase", "false", false},
		{"Valid false uppercase", "FALSE", false},
		{"Valid false mixedcase", "FaLsE", false},
		{"Numeric 1 as true", "1", true},
		{"Numeric 0 as false", "0", false},
		{"Invalid string", "abc", false}, // expect false because ParseBool fails
		{"Empty string", "", false},      // default false
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToBool(tt.input)
			if got != tt.expected {
				t.Errorf("ToBool(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
func TestRemoveAccents(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Cà phê sữa đá", "Ca phe sua da"},
		{"Thành phố Hồ Chí Minh", "Thanh pho Ho Chi Minh"},
		{"Điện Biên Phủ", "Dien Bien Phu"},
		{"Tôi yêu tiếng Việt", "Toi yeu tieng Viet"},
		{"Không dấu", "Khong dau"},
		{"Ñandú", "Nandu"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RemoveAccents(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeToASCII(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeText(t *testing.T) {
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
			result := Normalize(tt.input)
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
			actual := Truncate(tt.input, tt.max)
			if actual != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, actual)
			}
		})
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		input    string
		count    int
		char     rune
		expected string
	}{
		{"abc", 3, '*', "***abc"},
		{"hello", 0, '-', "hello"},
		{"go", -1, '!', "go"},
		{"", 2, '0', "00"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := PadLeft(tt.input, tt.count, tt.char)
			if result != tt.expected {
				t.Errorf("PadLeft(%q, %d, %q) = %q; want %q", tt.input, tt.count, tt.char, result, tt.expected)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		count    int
		char     rune
		expected string
	}{
		{"abc", 3, '*', "abc***"},
		{"hello", 0, '-', "hello"},
		{"go", -1, '!', "go"},
		{"", 2, '0', "00"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := PadRight(tt.input, tt.count, tt.char)
			if result != tt.expected {
				t.Errorf("PadRight(%q, %d, %q) = %q; want %q", tt.input, tt.count, tt.char, result, tt.expected)
			}
		})
	}
}

func TestPadCenter(t *testing.T) {
	tests := []struct {
		input    string
		start    int
		count    int
		char     rune
		expected string
	}{
		{"abcdef", 2, 3, '*', "ab***cdef"},
		{"end", 10, 3, '-', "end---"},
		{"hello", -5, 2, '_', "__hello"},
		{"", 0, 4, '#', "####"},
		{"12345", 2, 0, '*', "12345"},
		{"123", 1, -3, '!', "123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := PadCenter(tt.input, tt.start, tt.count, tt.char)
			if result != tt.expected {
				t.Errorf("PadCenter(%q, %d, %d, %q) = %q; want %q",
					tt.input, tt.start, tt.count, tt.char, result, tt.expected)
			}
		})
	}
}

func TestCompileRegex(t *testing.T) {
	pattern := `\d+`
	re, err := CompileRegex(pattern)
	if err != nil {
		t.Fatalf("unexpected error compiling pattern: %v", err)
	}
	matches := re.FindAllString("abc123def456", -1)
	expected := []string{"123", "456"}
	if len(matches) != len(expected) {
		t.Fatalf("expected %d matches, got %d", len(expected), len(matches))
	}
	for i := range matches {
		if matches[i] != expected[i] {
			t.Errorf("expected match %q, got %q", expected[i], matches[i])
		}
	}
}

func TestCompileRegex_InvalidPattern(t *testing.T) {
	_, err := CompileRegex(`(\`)
	if err == nil {
		t.Fatal("expected error compiling invalid pattern, got nil")
	}
}

func TestFindAllMatches(t *testing.T) {
	input := "Emails: test1@mail.com, test2@abc.org"
	pattern := `[\w\.-]+@[\w\.-]+\.\w+`
	expected := []string{"test1@mail.com", "test2@abc.org"}

	result := FindAllMatches(input, pattern)
	if len(result) != len(expected) {
		t.Fatalf("Expected %d matches, got %d", len(expected), len(result))
	}
	for i, match := range result {
		if match != expected[i] {
			t.Errorf("Expected %q, got %q", expected[i], match)
		}
	}
}

func TestContains(t *testing.T) {
	if !Contains("Hello, world", "world") {
		t.Error(`Contains("Hello, world", "world") = false; want true`)
	}
	if Contains("Hello, world", "mars") {
		t.Error(`Contains("Hello, world", "mars") = true; want false`)
	}
}

func TestStartWith(t *testing.T) {
	// positive
	if !StartWith("   Hello, world", "Hello") {
		t.Error(`StartWith("Hello, world", "Hello") = false; want true`)
	}

	// has white space or tab
	if !StartWith("   Hello, world", "Hello") {
		t.Error(`StartWith("Hello, world", "Hello") = false; want true`)
	}

	// has break line
	if !StartWith("\t\nHello, world", "Hello") {
		t.Error(`StartWith("Hello, world", "Hello") = false; want true`)
	}

	// negative
	if StartWith("Hello, world", "world") {
		t.Error(`StartWith("Hello, world", "world") = true; want false`)
	}
}

func TestEndWith(t *testing.T) {
	// positive
	if !EndWith("Hello, world", "world") {
		t.Error(`EndWith("Hello, world", "world") = false; want true`)
	}

	// has white space or tab
	if !EndWith("Hello, world    ", "world") {
		t.Error(`EndWith("Hello, world    ", "world") = false; want true`)
	}

	// has break line
	if !EndWith("Hello, world\n\t", "world") {
		t.Error(`EndWith("Hello, world\n\t", "world") = false; want true`)
	}

	// negative
	if EndWith("Hello, world", "Hello") {
		t.Error(`EndWith("Hello, world", "Hello") = true; want false`)
	}
}
