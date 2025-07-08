package validate

import (
	"context"
	"errors"
	"github.com/BevisDev/godev/consts"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// IsNilOrEmpty checks whether the given input is nil or empty.
//
// It supports multiple types including:
//   - nil pointers
//   - empty strings (after trimming spaces)
//   - empty arrays, slices, maps, or channels
//
// For all other types, it returns false by default.
//
// Examples:
//
//	IsNilOrEmpty(nil)                         // true
//	IsNilOrEmpty("")                          // true
//	IsNilOrEmpty("   ")                       // true
//	IsNilOrEmpty([]int{})                     // true
//	IsNilOrEmpty(map[string]string{})         // true
//	IsNilOrEmpty(123)                         // false
func IsNilOrEmpty(inp interface{}) bool {
	if inp == nil {
		return true
	}
	v := reflect.ValueOf(inp)

	// get val ptr
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	default:
		return false
	}
}

// IsNilOrZero checks whether a value is nil or equals to numeric zero.
// It supports int, int64, float64, uint, pointers to those types.
func IsNilOrZero(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)

	// Check if pointer and nil
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	default:
		return false
	}
}

func IsErrorOrEmpty(err error, i interface{}) bool {
	if err != nil || IsNilOrEmpty(i) {
		return true
	}
	return false
}

func IsPtr(i interface{}) bool {
	if i == nil {
		return false
	}
	v := reflect.ValueOf(i)
	return v.Kind() == reflect.Ptr && !v.IsNil()
}

func IsStruct(i interface{}) bool {
	if i == nil {
		return false
	}
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	return v.Kind() == reflect.Struct
}

func IsTimedOut(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

func IsValidPhoneNumber(phone string, size int) bool {
	length := len(phone)
	if length != size {
		return false
	}
	for _, r := range phone {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// Matches checks whether the input string `s` matches the given regular expression pattern.
//
// If the pattern is invalid, it safely returns false.
//
// Example:
//
//	Matches("hello123", "hello\\d+") // true
//	Matches("hello", "world")        // false
//	Matches("test", "(")             // false (invalid pattern)
//
// Parameters:
//   - s:      The input string to test.
//   - pattern: The regular expression pattern to match.
//
// Returns:
//   - true if the string matches the pattern; false otherwise.
func Matches(s, pattern string) bool {
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}
	return matched
}

func IsEmail(s string) bool {
	return Matches(s, consts.Email)
}

func IsPhoneNumber(s string) bool {
	return Matches(s, consts.TenDigitPhone)
}

func IsUUID(s string) bool {
	return Matches(s, consts.UUID)
}

func IsDate(s string) bool {
	return Matches(s, consts.DateYYYYMMDD)
}

func IsIPv4(s string) bool {
	matched := Matches(s, consts.IPv4)
	if !matched {
		return false
	}

	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || n > 255 {
			return false
		}
	}
	return true
}

func IsAlphaNumeric(s string) bool {
	return Matches(s, consts.AlphaNumeric)
}

func IsVietnamID(s string) bool {
	return Matches(s, consts.VNIDNumber)
}

func IsStrongPassword(s string, size int) bool {
	if len(s) < size {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r), unicode.IsSymbol(r):
			hasSpecial = true
		}

		if hasUpper && hasLower && hasDigit && hasSpecial {
			return true
		}
	}

	return false
}

// IsValidFileName checks whether a filename matches a predefined pattern
// and has an extension included in the allowed list.
//
// It performs two validations:
//  1. The filename structure matches the FilePattern regular expression.
//  2. The file extension (the part after the last dot) matches one of the allowed extensions (case-insensitive).
//
// If the filename does not contain a dot, or ends with a dot, or does not match the pattern, it is considered invalid.
//
// Example:
//
//	IsValidFileName("report.csv", []string{"csv", "txt"}) // true
//	IsValidFileName("archive.7z", []string{"zip"})       // false
//	IsValidFileName("badfile.", []string{"txt"})         // false
//	IsValidFileName("file", []string{"txt"})             // false
//
// Parameters:
//   - s: the filename to validate.
//   - allowedExt: a list of allowed extensions (without leading dots).
//
// Returns:
//   - true if the filename is valid and has an allowed extension.
//   - false otherwise.
func IsValidFileName(s string, allowedExt []string) bool {
	// Check pattern without enforcing extension length
	if !Matches(s, consts.FilePattern) {
		return false
	}

	// Extract extension
	lastDot := strings.LastIndex(s, ".")
	if lastDot == -1 || lastDot == len(s)-1 {
		return false
	}
	ext := s[lastDot+1:]

	// Compare (case-insensitive)
	extLower := strings.ToLower(ext)
	for _, allowed := range allowedExt {
		if extLower == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}
