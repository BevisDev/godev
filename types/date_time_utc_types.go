package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BevisDev/godev/utils/datetime"
)

type DateTimeUTC struct {
	time.Time
}

const layoutDateTimeUTC = datetime.DatetimeUTC

func (d *DateTimeUTC) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = DateTimeUTC{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := datetime.ToTime(s, layoutDateTimeUTC)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *DateTimeUTC) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(layoutDateTimeUTC))
}

func (d *DateTimeUTC) ToTime() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *DateTimeUTC) ToString() string {
	if d == nil || d.Time.IsZero() {
		return ""
	}
	return datetime.ToString(d.Time, layoutDateTimeUTC)
}

func (d *DateTimeUTC) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := datetime.ToTime(v, layoutDateTimeUTC)
		if err != nil {
			return fmt.Errorf("scan string to DateTimeUTC failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := datetime.ToTime(string(v), layoutDateTimeUTC)
		if err != nil {
			return fmt.Errorf("scan []byte to DateTimeUTC failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for DateTimeUTC.Scan: %T", v)
	}
	return nil
}
