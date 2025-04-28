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
