package helper

import (
	"reflect"
	"time"
)

func MapStructs(src interface{}, dest interface{}) {
	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(dest).Elem()
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		destField := destVal.Field(i)
		if srcField.Kind() == reflect.String && srcField.String() == "" {
			continue
		}
		if srcField.Type() == reflect.TypeOf(time.Time{}) && srcField.Interface().(time.Time).IsZero() {
			continue
		}
		if destField.Kind() == reflect.Ptr {
			if !srcField.IsZero() {
				destField.Set(reflect.New(destField.Type().Elem()))
				destField.Elem().Set(srcField)
			}
		} else {
			destField.Set(srcField)
		}
	}
}
