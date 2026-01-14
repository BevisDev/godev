package datetime

import (
	"database/sql/driver"
	"time"
)

type DBTime struct {
	baseTime
}

func (d *DBTime) IsZero() bool {
	return d.isZero()
}

func (d *DBTime) UnmarshalJSON(b []byte) error {
	return d.unmarshalLayout(b, DateTimeMillis)
}

func (d *DBTime) MarshalJSON() ([]byte, error) {
	return d.marshalLayout(DateTimeMillis)
}

func (d *DBTime) ToTime() *time.Time {
	t := d.Time
	return &t
}

func (d *DBTime) String() string {
	return d.stringLayout(DateTimeMillis)
}

func (d *DBTime) Scan(value interface{}) error {
	return d.scanLayout(value, DateTimeMillis)
}

func (d *DBTime) Value() (driver.Value, error) {
	return d.valueLayout(DateTimeMillis)
}
