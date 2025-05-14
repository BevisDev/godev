package str

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func ToString(value any) string {
	if value == nil {
		return ""
	}
	val := reflect.ValueOf(value)
	// handle ptr
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
	default:
		return fmt.Sprintf("%+v", val.Interface())
	}
	return fmt.Sprintf("%+v", val.Interface())
}

func ToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func ToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func ToFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return f
}

// NormalizeToASCII converts a Unicode string to its ASCII equivalent.
//
// It removes diacritical marks (e.g., accents), and replaces Vietnamese
// characters "Đ"/"đ" with "D"/"d" explicitly.
//
// Example:
//
//	NormalizeToASCII("Bình Đẹp Trai") → "Binh Dep Trai"
func NormalizeToASCII(str string) string {
	result := norm.NFD.String(str)
	var output []rune
	for _, r := range result {
		if unicode.Is(unicode.M, r) {
			continue
		}
		output = append(output, r)
	}
	normalized := string(output)
	normalized = strings.ReplaceAll(normalized, "Đ", "D")
	normalized = strings.ReplaceAll(normalized, "đ", "d")
	return normalized
}

// CleanText normalizes a string to ASCII and removes all non-alphanumeric characters,
// except spaces.
//
// It is useful for generating slugs, sanitized input, or matching keywords.
//
// Example:
//
//	CleanText("Đặng Văn Lâm!!!") → "Dang Van Lam"
func CleanText(str string) string {
	o := NormalizeToASCII(str)
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

// TruncateText limits the length of a string to at most `maxLen` runes.
// If the input string is shorter than or equal to `maxLen`, it is returned unchanged.
//
// Example:
//
//	TruncateText("Hello, world!", 5) → "Hello"
func TruncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

func PadLeft(s string, count int, c rune) string {
	if count <= 0 {
		return s
	}
	return strings.Repeat(string(c), count) + s
}

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
