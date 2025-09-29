package jsonx

import (
	"encoding/json"
	"errors"
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
	return string(jsonBytes)
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

// Clone clones any struct or map via JSON marshal/unmarshal.
// Note: Won't work if the struct has unexported fields.
func Clone[T any](src T) (T, error) {
	var dst T
	b, err := json.Marshal(src)
	if err != nil {
		return dst, err
	}

	err = json.Unmarshal(b, &dst)
	return dst, err
}

func StructToMap(entry interface{}) map[string]interface{} {
	j := ToJSONBytes(entry)
	var result map[string]interface{}
	if err := JSONBytesToStruct(j, &result); err != nil {
		return nil
	}
	return result
}

func Pretty(v any) string {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "{}"
	}
	return string(b)
}
