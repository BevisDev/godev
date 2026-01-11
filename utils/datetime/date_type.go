package datetime

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) IsZero() bool {
	return d == nil || d.Time.IsZero()
}

func (d *Date) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = Date{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := ToTime(s, DateLayoutISO)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format(DateLayoutISO))
}

func (d *Date) ToTime() *time.Time {
	if d.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *Date) String() string {
	if d.IsZero() {
		return ""
	}
	return ToString(d.Time, DateLayoutISO)
}

func (d *Date) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := ToTime(v, DateLayoutISO)
		if err != nil {
			return fmt.Errorf("scan string to Date failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := ToTime(string(v), DateLayoutISO)
		if err != nil {
			return fmt.Errorf("scan []byte to Date failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for Date.Scan: %T", v)
	}
	return nil
}

func (d *Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Format(DateLayoutISO), nil
}
