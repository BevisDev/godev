package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BevisDev/godev/utils/datetime"
)

type Date struct {
	time.Time
}

const layoutDate = datetime.DateOnly

func (d *Date) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = Date{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := datetime.ToTime(s, layoutDate)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(layoutDate))
}

func (d *Date) ToTime() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *Date) ToString() string {
	if d == nil || d.Time.IsZero() {
		return ""
	}
	return datetime.ToString(d.Time, layoutDate)
}

func (d *Date) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := datetime.ToTime(v, layoutDate)
		if err != nil {
			return fmt.Errorf("scan string to Date failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := datetime.ToTime(string(v), layoutDate)
		if err != nil {
			return fmt.Errorf("scan []byte to Date failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for Date.Scan: %T", v)
	}
	return nil
}
