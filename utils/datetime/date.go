package datetime

import (
	"database/sql/driver"
	"time"
)

type Date struct {
	baseTime
}

// NewDate returns current date time.
func NewDate() *Date {
	return &Date{
		baseTime: baseTime{
			Time: time.Now(),
		},
	}
}

// ToDate parses a date string into Date using the specified layout.
//
// Example:
//
//	d, err := ToDate("2024-01-02")
func ToDate(str string) (*Date, error) {
	parsedTime, err := ToTime(str, DateLayoutISO)
	if err != nil {
		return nil, err
	}

	return &Date{
		baseTime: baseTime{
			Time: *parsedTime,
		},
	}, nil
}

func (d *Date) IsZero() bool {
	return d.isZero()
}

func (d *Date) UnmarshalJSON(b []byte) error {
	return d.unmarshalLayout(b, DateLayoutISO)
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return d.marshalLayout(DateLayoutISO)
}

func (d *Date) ToTime() *time.Time {
	t := d.Time
	return &t
}

func (d *Date) String() string {
	return d.stringLayout(DateLayoutISO)
}

func (d *Date) Scan(value interface{}) error {
	return d.scanLayout(value, DateLayoutISO)
}

func (d *Date) Value() (driver.Value, error) {
	return d.valueLayout(DateLayoutISO)
}
