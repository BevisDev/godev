package validate

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
