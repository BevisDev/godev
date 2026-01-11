package datetime

import (
	"encoding/json"
	"fmt"
	"time"
)

type UTCTime struct {
	time.Time
}

func (d *UTCTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = UTCTime{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := ToTime(s, DateTimeLayoutUTC)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *UTCTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DateTimeLayoutUTC))
}

func (d *UTCTime) ToTime() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *UTCTime) ToString() string {
	if d == nil || d.Time.IsZero() {
		return ""
	}
	return ToString(d.Time, DateTimeLayoutUTC)
}

func (d *UTCTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := ToTime(v, DateTimeLayoutUTC)
		if err != nil {
			return fmt.Errorf("scan string to UTCTime failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := ToTime(string(v), DateTimeLayoutUTC)
		if err != nil {
			return fmt.Errorf("scan []byte to UTCTime failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for UTCTime.Scan: %T", v)
	}
	return nil
}
