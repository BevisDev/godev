package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BevisDev/godev/utils/datetime"
)

type DateTime struct {
	time.Time
}

func (d *DateTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = DateTime{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := datetime.ToTime(s, datetime.DateTimeNoTZ)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(datetime.DateTimeNoTZ))
}

func (d *DateTime) ToTime() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *DateTime) ToString() string {
	if d == nil || d.Time.IsZero() {
		return ""
	}
	return datetime.ToString(d.Time, datetime.DateTimeNoTZ)
}

func (d *DateTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := datetime.ToTime(v, datetime.DateTimeNoTZ)
		if err != nil {
			return fmt.Errorf("scan string to DateTime failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := datetime.ToTime(string(v), datetime.DateTimeNoTZ)
		if err != nil {
			return fmt.Errorf("scan []byte to DateTime failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for DateTime.Scan: %T", v)
	}
	return nil
}
