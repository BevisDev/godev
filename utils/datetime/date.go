package datetime

import (
	"database/sql/driver"
	"time"
)

type Date struct {
	baseTime
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
