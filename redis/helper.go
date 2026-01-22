package redis

import (
	"encoding/json"
	"fmt"
)

// convertValue converts a value to a format suitable for Redis storage.
// Strings and bytes are kept as-is, primitives are converted to strings,
// and complex types are JSON-marshaled.
func convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return v
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return fmt.Sprint(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return b
	}
}
