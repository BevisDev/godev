package str

import (
	"encoding/json"
	"fmt"
	"github.com/BevisDev/godev/types"
	"github.com/shopspring/decimal"
	"golang.org/x/text/unicode/norm"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ToString converts a value of any supported type to its string representation.
//
// It handles common primitive types including:
//   - string
//   - int, uint (and their variants)
//   - float32, float64
//   - bool
//   - []byte (converted to string)
//
// If the value is a pointer, it will be dereferenced (unless nil).
// For unsupported or complex types, it falls back to fmt.Sprintf("%+v", value).
//
// Returns an empty string if the value is nil or a nil pointer.
//
// Examples:
//
//	ToString(123)          → "123"
//	ToString(3.14)         → "3.14"
//	ToString(true)         → "true"
//	ToString([]byte("hi")) → "hi"
//	ToString(nil)          → ""
func ToString(value any) string {
	if value == nil {
		return ""
	}
	val := reflect.ValueOf(value)

	// get val ptr
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(val.Float(), 'g', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return string(val.Bytes()) // handle []byte
		}
	case reflect.Struct:
		if t, ok := val.Interface().(time.Time); ok {
			return t.Format(time.RFC3339)
		}
		if d, ok := val.Interface().(decimal.Decimal); ok {
			return d.String()
		}
		if b, err := json.Marshal(val.Interface()); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%+v", val.Interface())
	default:
		return fmt.Sprintf("%+v", val.Interface())
	}
	return fmt.Sprintf("%+v", val.Interface())
}

// ToInt parses a string to any signed integer type (int, int16, ...).
func ToInt[T types.SignedInteger](s string) T {
	var zero T
	var bitSize int

	switch any(zero).(type) {
	case int:
		i, err := strconv.Atoi(s)
		if err != nil {
			return zero
		}
		return T(i)
	case int8:
		bitSize = 8
	case int16:
		bitSize = 16
	case int32:
		bitSize = 32
	case int64:
		bitSize = 64
	default:
		return zero
	}

	i, err := strconv.ParseInt(s, 10, bitSize)
	if err != nil {
		return zero
	}
	return T(i)
}

// ToFloat parses a string to float32 or float64.
func ToFloat[T types.SignedFloat](str string) T {
	var zero T
	var bitSize int

	switch any(zero).(type) {
	case float32:
		bitSize = 32
	case float64:
		bitSize = 64
	default:
		return zero
	}

	f, err := strconv.ParseFloat(str, bitSize)
	if err != nil {
		return zero
	}
	return T(f)
}

// RemoveAccents converts a Unicode string to its equivalent without diacritical marks.
//
// It removes accents and diacritical marks (e.g., á → a, ü → u) and explicitly
// replaces Vietnamese characters "Đ"/"đ" with "D"/"d".
//
// Example:
//
//	RemoveAccents("Đặng Thị Ánh")   // "Dang Thi Anh"
//	RemoveAccents("Café Noël ü")   // "Cafe Noel u"
func RemoveAccents(str string) string {
	if str == "" {
		return ""
	}

	// Decompose characters
	decomposed := norm.NFD.String(str)

	var builder strings.Builder
	builder.Grow(len(decomposed))

	for _, r := range decomposed {
		if unicode.Is(unicode.M, r) {
			continue
		}

		// replace directly
		switch r {
		case 'Đ':
			builder.WriteRune('D')
		case 'đ':
			builder.WriteRune('d')
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// Clean normalizes a string to ASCII and removes all non-alphanumeric characters,
// except spaces.
//
// It is useful for generating slugs, sanitized input, or matching keywords.
//
// Example:
//
//	Clean("Đặng Thị Ánh ♥ 123!") → "Dang Thi Anh 123"
func Clean(str string) string {
	o := RemoveAccents(str)
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	return re.ReplaceAllString(o, "")
}

// RemoveWhiteSpace removes all space characters from a string.
//
// Example:
//
//	RemoveWhiteSpace("a b c") → "abc"
func RemoveWhiteSpace(str string) string {
	return strings.ReplaceAll(str, " ", "")
}

// Truncate limits the length of a string to at most `maxLen` runes.
// If the input string is shorter than or equal to `maxLen`, it is returned unchanged.
//
// Example:
//
//	Truncate("Hello, world!", 5) → "Hello"
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

// PadLeft prepends the rune `c` to the string `s` `count` times.
//
// If `count` is less than or equal to 0, it returns `s` unchanged.
//
// Example:
//
//	PadLeft("abc", 3, '0') => "000abc"
//	PadLeft("hello", 3, 'A') => "AAAhello"
func PadLeft(s string, count int, c rune) string {
	if count <= 0 {
		return s
	}
	return strings.Repeat(string(c), count) + s
}

// PadRight appends the rune `c` to the string `s` `count` times.
//
// If `count` is less than or equal to 0, it returns `s` unchanged.
//
// Example:
//
//	PadRight("abc", 3, '0') => "abc000"
//	PadRight("abc", 3, 'A') => "abcAAA"
func PadRight(s string, count int, c rune) string {
	if count <= 0 {
		return s
	}
	return s + strings.Repeat(string(c), count)
}

// PadCenter inserts `count` copies of the character `c` into the string `s`
// starting at the specified `start` index.
//
// If `start` is less than 0, it defaults to 0.
// If `start` is greater than len(s), it defaults to the end of the string.
// If `count` is less than or equal to 0, the original string is returned.
//
// Example:
//
//	fmt.Println(PadCenter("abcdef", 2, 3, '*'))
//	fmt.Println(PadCenter("end", 10, 3, '-'))
//	fmt.Println(PadCenter("hello", -5, 2, '_'))
//
// Output:
//
//	ab***cdef
//	end---
//	__hello
func PadCenter(s string, start int, count int, c rune) string {
	if count <= 0 {
		return s
	}

	if start < 0 {
		start = 0
	}
	if start > len(s) {
		start = len(s)
	}

	insert := strings.Repeat(string(c), count)
	return s[:start] + insert + s[start:]
}

// CompileRegex compiles a regular expression pattern into a Regexp object.
//
// It returns an error if the pattern is invalid.
//
// Example:
//
//	re, err := CompileRegex(`\d+`)
//	if err != nil {
//	    // handle error
//	}
//	matches := re.FindAllString("abc123def456", -1) // ["123", "456"]
func CompileRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// FindAllMatches finds all non-overlapping matches of the regular expression pattern
// in the input string s and returns them as a slice of strings.
//
// If the pattern is invalid, it returns nil.
//
// Example:
//
//	FindAllMatches("a1b2c3", `\d`) → []string{"1", "2", "3"}
func FindAllMatches(s, pattern string) []string {
	re, err := CompileRegex(pattern)
	if err != nil {
		return nil
	}
	return re.FindAllString(s, -1)
}

// Contains reports whether the substring subStr is within s.
//
// Example:
//
//	Contains("Hello, world", "world") → true
func Contains(s, subStr string) bool {
	return strings.Contains(s, subStr)
}

// StartWith reports whether the string s begins with substring subStr.
//
// Example:
//
//	StartWith("Hello, world", "Hello") → true
func StartWith(s, subStr string) bool {
	return strings.HasPrefix(s, subStr)
}

// EndWith reports whether the string s ends with substring subStr.
//
// Example:
//
//	EndWith("Hello, world", "world") → true
func EndWith(s, subStr string) bool {
	return strings.HasSuffix(s, subStr)
}
