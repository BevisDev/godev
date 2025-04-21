package types

import (
	"encoding/json"
	"fmt"
	"github.com/BevisDev/godev/utils/datetime"
	"strings"
	"time"
)

type DateSQL struct {
	time.Time
}

func (d *DateSQL) UnmarshalJSON(b []byte) error {
	str := strings.TrimSpace(string(b))
	if str == "null" {
		*d = DateSQL{}
		return nil
	}

	s := strings.Trim(str, `"`)
	t, err := datetime.ToTime(s, datetime.DateTimeSQL)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *DateSQL) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(datetime.DateTimeSQL))
}

func (d *DateSQL) ToTime() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *DateSQL) ToString() string {
	if d == nil || d.Time.IsZero() {
		return ""
	}
	return datetime.ToString(d.Time, datetime.DateTimeSQL)
}

func (d *DateSQL) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := datetime.ToTime(v, datetime.DateTimeSQL)
		if err != nil {
			return fmt.Errorf("scan string to DateSQL failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := datetime.ToTime(string(v), datetime.DateTimeSQL)
		if err != nil {
			return fmt.Errorf("scan []byte to DateSQL failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for DateSQL.Scan: %T", v)
	}
	return nil
}
