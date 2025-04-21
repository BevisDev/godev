package types

import (
	"encoding/json"
	"fmt"
	"github.com/BevisDev/godev/utils/datetime"
	"strings"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(b []byte) error {
	str := strings.TrimSpace(string(b))
	if str == "null" {
		*d = Date{}
		return nil
	}

	s := strings.Trim(str, `"`)
	t, err := datetime.ToTime(s, datetime.DateOnly)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(datetime.DateOnly))
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
	return datetime.ToString(d.Time, datetime.DateOnly)
}

func (d *Date) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := datetime.ToTime(v, datetime.DateOnly)
		if err != nil {
			return fmt.Errorf("scan string to Date failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := datetime.ToTime(string(v), datetime.DateOnly)
		if err != nil {
			return fmt.Errorf("scan []byte to Date failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for Date.Scan: %T", v)
	}
	return nil
}
