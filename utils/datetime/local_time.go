package datetime

import (
	"database/sql/driver"
	"time"
)

// LocalTime represents a datetime without timezone information.
// It is typically used for business logic and DB timestamps without TZ.
type LocalTime struct {
	baseTime
}

// NewLocalTime returns current local datetime.
func NewLocalTime() *LocalTime {
	return &LocalTime{
		baseTime: baseTime{
			Time: time.Now(),
		},
	}
}

// ToLocalTime parses a datetime string into LocalTime using the local datetime layout.
//
// Example:
//
//	d, err := ToLocalTime("2024-01-02T15:04:05")
func ToLocalTime(str string) (*LocalTime, error) {
	parsedTime, err := ToTime(str, DateTimeLayoutLocal)
	if err != nil {
		return nil, err
	}

	return &LocalTime{
		baseTime: baseTime{
			Time: *parsedTime,
		},
	}, nil
}

func (d *LocalTime) IsZero() bool {
	return d.isZero()
}

func (d *LocalTime) UnmarshalJSON(b []byte) error {
	return d.unmarshalLayout(b, DateTimeLayoutLocal)
}

func (d *LocalTime) MarshalJSON() ([]byte, error) {
	return d.marshalLayout(DateTimeLayoutLocal)
}

func (d *LocalTime) ToTime() *time.Time {
	t := d.Time
	return &t
}

func (d *LocalTime) String() string {
	return d.stringLayout(DateTimeLayoutLocal)
}

func (d *LocalTime) Scan(value interface{}) error {
	return d.scanLayout(value, DateTimeLayoutLocal)
}

func (d *LocalTime) Value() (driver.Value, error) {
	return d.valueLayout(DateTimeLayoutLocal)
}
