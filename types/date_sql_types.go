package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BevisDev/godev/utils/datetime"
)

type DateSQL struct {
	time.Time
}

const layoutDateSQL = datetime.DateTimeSQL

func (d *DateSQL) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*d = DateSQL{}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("invalid JSON string: %w", err)
	}

	t, err := datetime.ToTime(s, layoutDateSQL)
	if err != nil {
		return err
	}

	d.Time = *t
	return nil
}

func (d *DateSQL) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(layoutDateSQL))
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
	return datetime.ToString(d.Time, layoutDateSQL)
}

func (d *DateSQL) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case string:
		t, err := datetime.ToTime(v, layoutDateSQL)
		if err != nil {
			return fmt.Errorf("scan string to DateSQL failed: %w", err)
		}
		d.Time = *t
	case []byte:
		t, err := datetime.ToTime(string(v), layoutDateSQL)
		if err != nil {
			return fmt.Errorf("scan []byte to DateSQL failed: %w", err)
		}
		d.Time = *t
	default:
		return fmt.Errorf("unsupported type for DateSQL.Scan: %T", v)
	}
	return nil
}
