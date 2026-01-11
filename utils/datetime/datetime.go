package datetime

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type DateTime struct {
	time.Time
}

func (d *DateTime) IsZero() bool {
	return d == nil || d.Time.IsZero()
}

func (d *DateTime) unmarshalWithLayout(b []byte, layout string) error {
	if string(b) == "null" {
		*d = DateTime{}
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

	d.Time = *t
	return nil
}

func (d *DateTime) marshalWithLayout(layout string) ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format(layout))
}

func (d *DateTime) stringWithLayout(layout string) string {
	if d.IsZero() {
		return ""
	}
	return ToString(d.Time, layout)
}

func (d *Date) scanWithLayout(value any, layout string) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := ToTime(v, layout)
		if err != nil {
			return fmt.Errorf("scan string to Date failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := ToTime(string(v), layout)
		if err != nil {
			return fmt.Errorf("scan []byte to Date failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for Date.Scan: %T", v)
	}
	return nil
}

func (d *Date) valueWithLayout(layout string) (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Format(layout), nil
}
