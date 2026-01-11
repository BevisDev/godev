package datetime

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type DBTime struct {
	time.Time
}

func (d *DBTime) IsZero() bool {
	return d == nil || d.Time.IsZero()
}

func (d *DBTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = DBTime{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := ToTime(s, DateTimeMillis)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *DBTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DateTimeMillis))
}

func (d *DBTime) ToTime() *time.Time {
	if d.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}

func (d *DBTime) String() string {
	if d.IsZero() {
		return ""
	}
	return ToString(d.Time, DateTimeMillis)
}

func (d *DBTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := ToTime(v, DateTimeMillis)
		if err != nil {
			return fmt.Errorf("scan string to DBTime failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := ToTime(string(v), DateTimeMillis)
		if err != nil {
			return fmt.Errorf("scan []byte to DBTime failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for DBTime.Scan: %T", v)
	}
	return nil
}

func (d *DBTime) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Format(DateTimeMillis), nil
}
