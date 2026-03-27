package datetime

import (
	"database/sql/driver"
	"time"
)

type UTCTime struct {
	baseTime
}

// NewUTCTime returns current datetime in UTC.
func NewUTCTime() *UTCTime {
	return &UTCTime{
		baseTime: baseTime{
			Time: time.Now().UTC(),
		},
	}
}

// ToUTCTime parses a datetime string into UTCTime using the UTC layout.
//
// Example:
//
//	d, err := ToUTCTime("2024-01-02T15:04:05Z")
func ToUTCTime(str string) (*UTCTime, error) {
	parsedTime, err := ToTime(str, DateTimeLayoutUTC)
	if err != nil {
		return nil, err
	}

	return &UTCTime{
		baseTime: baseTime{
			Time: *parsedTime,
		},
	}, nil
}

func (d *UTCTime) IsZero() bool {
	return d.isZero()
}

func (d *UTCTime) UnmarshalJSON(b []byte) error {
	return d.unmarshalLayout(b, DateTimeLayoutUTC)
}

func (d *UTCTime) MarshalJSON() ([]byte, error) {
	return d.marshalLayout(DateTimeLayoutUTC)
}

func (d *UTCTime) ToTime() *time.Time {
	t := d.Time
	return &t
}

func (d *UTCTime) String() string {
	return d.stringLayout(DateTimeLayoutUTC)
}

func (d *UTCTime) Scan(value interface{}) error {
	return d.scanLayout(value, DateTimeLayoutUTC)
}

func (d *UTCTime) Value() (driver.Value, error) {
	return d.valueLayout(DateTimeLayoutUTC)
}
