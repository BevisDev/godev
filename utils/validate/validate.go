package validate

import (
	"context"
	"errors"
	"reflect"
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
