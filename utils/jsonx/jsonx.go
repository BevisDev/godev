package jsonx

import (
	"encoding/json"

	"github.com/BevisDev/godev/utils/validate"
)

// ToJSONBytes marshals a value into JSON bytes.
func ToJSONBytes(v any) ([]byte, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

// FromJSONBytes unmarshals JSON bytes into type T.
func FromJSONBytes[T any](raw []byte) (T, error) {
	var t T
	err := json.Unmarshal(raw, &t)
	return t, err
}

// FromJSON unmarshals a JSON string into type T.
func FromJSON[T any](str string) (T, error) {
	return FromJSONBytes[T]([]byte(str))
}

// ToJSON marshals a value into a JSON string.
func ToJSON(v any) string {
	jsonBytes, err := ToJSONBytes(v)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

// ObjectToMap converts a struct value into a map.
func ObjectToMap(i interface{}) map[string]interface{} {
	if !validate.IsStruct(i) {
		return nil
	}

	raw, err := ToJSONBytes(i)
	if err != nil {
		return nil
	}

	out, err := FromJSONBytes[map[string]interface{}](raw)
	if err != nil {
		return nil
	}
	return out
}

// JSONStringToMap converts a JSON string into a map.
func JSONStringToMap(s string) (map[string]interface{}, error) {
	return FromJSON[map[string]interface{}](s)
}

// JSONBytesToMap converts JSON bytes into a map.
func JSONBytesToMap(b []byte) (map[string]interface{}, error) {
	return FromJSONBytes[map[string]interface{}](b)
}

// Clone deep-copies a value via JSON marshal/unmarshal.
func Clone[T any](src T) (T, error) {
	var dst T
	b, err := ToJSONBytes(src)
	if err != nil {
		return dst, err
	}
	return FromJSONBytes[T](b)
}

// Pretty returns an indented JSON string for a value.
func Pretty(v any) string {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "{}"
	}
	return string(b)
}
