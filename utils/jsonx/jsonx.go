package jsonx

import (
	"encoding/json"
	"errors"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
)

func ToJSONBytes(v any) []byte {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return jsonBytes
}

func JSONBytesToStruct(jsonBytes []byte, entry interface{}) error {
	if !validate.IsPtr(entry) {
		return errors.New("must be a pointer")
	}
	err := json.Unmarshal(jsonBytes, entry)
	if err != nil {
		return err
	}
	return nil
}

func ToJSON(v any) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return str.ToString(jsonBytes)
}

func ToStruct(jsonStr string, entry interface{}) error {
	if !validate.IsPtr(entry) {
		return errors.New("must be a pointer")
	}
	err := json.Unmarshal([]byte(jsonStr), entry)
	if err != nil {
		return err
	}
	return nil
}

func StructToMap(entry interface{}) map[string]interface{} {
	j := ToJSONBytes(entry)
	var result map[string]interface{}
	if err := JSONBytesToStruct(j, &result); err != nil {
		return nil
	}
	return result
}

func Pretty[T any](v T) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
