package jsonx

import (
	"encoding/json"
)

func ToJSONBytes(v any) ([]byte, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func FromJSONBytes[T any](raw []byte) (T, error) {
	var t T
	err := json.Unmarshal(raw, &t)
	return t, err
}

func FromJSON[T any](str string) (T, error) {
	return FromJSONBytes[T]([]byte(str))
}

func ToJSON(v any) string {
	jsonBytes, err := ToJSONBytes(v)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

func StructToMap(i interface{}) map[string]interface{} {
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

// Clone clones any struct or map via JSON marshal/unmarshal.
// Note: Won't work if the struct has unexported fields.
func Clone[T any](src T) (T, error) {
	var dst T
	b, err := ToJSONBytes(src)
	if err != nil {
		return dst, err
	}
	return FromJSONBytes[T](b)
}

func Pretty(v any) string {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "{}"
	}
	return string(b)
}
