package helper

import (
	"encoding/json"
)

func ToJSON(v any) []byte {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return jsonBytes
}

func JSONToStr(jsonBytes []byte) string {
	return string(jsonBytes)
}

func ToJSONStr(v any) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return JSONToStr(jsonBytes)
}

func ToStruct(jsonBytes []byte, result any) error {
	err := json.Unmarshal(jsonBytes, result)
	if err != nil {
		return err
	}
	return nil
}

func JSONToStruct(jsonStr string, result any) error {
	err := json.Unmarshal([]byte(jsonStr), result)
	if err != nil {
		return err
	}
	return nil
}
