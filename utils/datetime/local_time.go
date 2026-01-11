package datetime

import (
	"encoding/json"
	"fmt"
	"time"
)

// LocalTime represents a datetime without timezone information.
// It is typically used for business logic and DB timestamps without TZ.
type LocalTime struct {
	time.Time
}

func (d *LocalTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = LocalTime{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := ToTime(s, DateTimeLayoutLocal)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *LocalTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DateTimeLayoutLocal))
}

func (d *LocalTime) ToTime() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *LocalTime) ToString() string {
	if d == nil || d.Time.IsZero() {
		return ""
	}
	return ToString(d.Time, DateTimeLayoutLocal)
}

func (d *LocalTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := ToTime(v, DateTimeLayoutLocal)
		if err != nil {
			return fmt.Errorf("scan string to LocalTime failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := ToTime(string(v), DateTimeLayoutLocal)
		if err != nil {
			return fmt.Errorf("scan []byte to LocalTime failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for LocalTime.Scan: %T", v)
	}
	return nil
}
