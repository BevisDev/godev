package helper

import (
	"encoding/json"
)

func ToJSONBytes(v any) []byte {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return jsonBytes
}

func JSONBytesToStruct(jsonBytes []byte, result any) error {
	err := json.Unmarshal(jsonBytes, result)
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

func JSONToStruct(jsonStr string, result any) error {
	err := json.Unmarshal([]byte(jsonStr), result)
	if err != nil {
		return err
	}
	return nil
}
