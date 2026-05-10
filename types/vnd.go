package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/BevisDev/godev/consts"
)

// VND Money type
type VND int64

// Zero is the zero value for VND.
const Zero VND = 0

func New(v int64) VND {
	return VND(v)
}

// FromString parses a VND value from a numeric string (e.g. "1000000").
// Commas and spaces are stripped before parsing.
func FromString(s string) (VND, error) {
	s = strings.ReplaceAll(s, consts.Comma, consts.Empty)
	s = strings.ReplaceAll(s, consts.Space, consts.Empty)
	s = strings.TrimSuffix(s, consts.VND)
	s = strings.TrimSpace(s)
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("vnd: cannot parse %q: %w", s, err)
	}
	return VND(v), nil
}

// Int64 returns the raw int64 value.
func (v VND) Int64() int64 {
	return int64(v)
}

// IsZero reports whether v is zero.
func (v VND) IsZero() bool {
	return v == Zero
}

// Percent returns p% of v (e.g. v.Percent(10) → 10%).
func (v VND) Percent(p int64) VND {
	return VND(int64(v) * p / 100)
}

// MarshalJSON encodes as a JSON number (e.g. 1000000).
func (v VND) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(v))
}

// UnmarshalJSON decodes from a JSON number or numeric string.
func (v *VND) UnmarshalJSON(data []byte) error {
	// Try number first
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*v = VND(n)
		return nil
	}
	// Fallback: quoted string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("vnd: cannot unmarshal %s", data)
	}
	parsed, err := FromString(s)
	if err != nil {
		return err
	}
	*v = parsed
	return nil
}

// Value implements driver.Valuer — stores as int64 in the DB.
func (v VND) Value() (driver.Value, error) {
	return int64(v), nil
}

// Scan implements sql.Scanner — reads int64 or []byte/string from the DB.
func (v *VND) Scan(src any) error {
	switch val := src.(type) {
	case int64:
		*v = VND(val)
	case float64:
		*v = VND(int64(val))
	case []byte:
		parsed, err := FromString(string(val))
		if err != nil {
			return err
		}
		*v = parsed
	case string:
		parsed, err := FromString(val)
		if err != nil {
			return err
		}
		*v = parsed
	case nil:
		*v = 0
	default:
		return fmt.Errorf("vnd: cannot scan type %T", src)
	}
	return nil
}
