package datetime

import (
	"database/sql/driver"
	"time"
)

type DBTime struct {
	baseTime
}

// NewDBTime returns current datetime.
func NewDBTime() *DBTime {
	return &DBTime{
		baseTime: baseTime{
			Time: time.Now(),
		},
	}
}

// ToDBTime parses a datetime string into DBTime using the DB layout.
//
// Example:
//
//	d, err := ToDBTime("2024-01-02 15:04:05.000")
func ToDBTime(str string) (*DBTime, error) {
	parsedTime, err := ToTime(str, DateTimeLayoutMilli)
	if err != nil {
		return nil, err
	}

	return &DBTime{
		baseTime: baseTime{
			Time: *parsedTime,
		},
	}, nil
}

func (d *DBTime) IsZero() bool {
	return d.isZero()
}

func (d *DBTime) UnmarshalJSON(b []byte) error {
	return d.unmarshalLayout(b, DateTimeLayoutMilli)
}

func (d *DBTime) MarshalJSON() ([]byte, error) {
	return d.marshalLayout(DateTimeLayoutMilli)
}

func (d *DBTime) ToTime() *time.Time {
	t := d.Time
	return &t
}

func (d *DBTime) String() string {
	return d.stringLayout(DateTimeLayoutMilli)
}

func (d *DBTime) Scan(value interface{}) error {
	return d.scanLayout(value, DateTimeLayoutMilli)
}

func (d *DBTime) Value() (driver.Value, error) {
	return d.valueLayout(DateTimeLayoutMilli)
}
