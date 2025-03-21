package helper

import (
	"context"
	"errors"
	"reflect"
	"strings"
)

func IsNilOrEmpty(inp interface{}) bool {
	if inp == nil {
		return true
	}
	v := reflect.ValueOf(inp)
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Array, reflect.Slice:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	case reflect.Chan:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

func IsContains[T comparable](arr []T, value T) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func IsTimedOut(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

func CompareStringWithAccent(str1, str2 string) bool {
	return strings.EqualFold(str1, str2)
}

func CompareStringWithoutAccent(str1, str2 string) bool {
	o1 := RemoveAccent(str1)
	o2 := RemoveAccent(str2)
	return strings.EqualFold(o1, o2)
}

func CompareStringWithoutWhitespace(str1, str2 string) bool {
	o1 := RemoveWhiteSpace(str1)
	o2 := RemoveWhiteSpace(str2)
	return strings.EqualFold(o1, o2)
}

func CompareStringWithoutSpecialChars(str1, str2 string) bool {
	o1 := RemoveSpecialChars(str1)
	o2 := RemoveSpecialChars(str2)
	return strings.EqualFold(o1, o2)
}
