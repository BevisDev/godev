package datetime

import (
	"database/sql/driver"
	"time"
)

type UTCTime struct {
	baseTime
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
