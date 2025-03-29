package utils

import (
	"encoding/json"
	"errors"
)

func ToJSONBytes(v any) []byte {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return jsonBytes
}

func JSONBytesToStruct(jsonBytes []byte, entry interface{}) error {
	if !IsPointer(entry) {
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
		return ""
	}
	return ToString(jsonBytes)
}

func JSONToStruct(jsonStr string, entry interface{}) error {
	if !IsPointer(entry) {
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
