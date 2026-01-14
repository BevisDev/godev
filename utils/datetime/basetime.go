package datetime

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type baseTime struct {
	time.Time
}

func (bt *baseTime) isZero() bool {
	return bt == nil || bt.Time.IsZero()
}

func (bt *baseTime) unmarshalLayout(b []byte, layout string) error {
	if string(b) == "null" {
		*bt = baseTime{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := ToTime(s, layout)
	if err != nil {
		return err
	}

	bt.Time = *t
	return nil
}

func (bt *baseTime) marshalLayout(layout string) ([]byte, error) {
	if bt.isZero() {
		return []byte("null"), nil
	}
	return json.Marshal(bt.Format(layout))
}

func (bt *baseTime) stringLayout(layout string) string {
	if bt.isZero() {
		return ""
	}
	return ToString(bt.Time, layout)
}

func (bt *baseTime) scanLayout(value any, layout string) error {
	if value == nil {
		*bt = baseTime{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		bt.Time = v
	case string:
		t, err := ToTime(v, layout)
		if err != nil {
			return fmt.Errorf("scan string to Date failed: %w", err)
		}
		bt.Time = *t
	case []byte:
		t, err := ToTime(string(v), layout)
		if err != nil {
			return fmt.Errorf("scan []byte to Date failed: %w", err)
		}
		bt.Time = *t
	default:
		return fmt.Errorf("unsupported type for scan: %T", v)
	}
	return nil
}

func (bt *baseTime) valueLayout(layout string) (driver.Value, error) {
	if bt.isZero() {
		return nil, nil
	}
	return bt.Format(layout), nil
}
