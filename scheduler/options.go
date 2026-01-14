package scheduler

import (
	"log"
	"time"
)

type OptionFunc func(*options)

type options struct {
	Location    *time.Location
	WithSeconds bool
}

func defaultOptions() *options {
	return &options{
		Location:    time.UTC,
		WithSeconds: false,
	}
}

func WithSeconds() OptionFunc {
	return func(o *options) {
		o.WithSeconds = true
	}
}

func WithLocation(loc *time.Location) OptionFunc {
	return func(o *options) {
		if loc != nil {
			o.Location = loc
		}
	}
}

func WithTimezone(tz string) OptionFunc {
	return func(o *options) {
		if tz == "" {
			o.Location = time.UTC
			return
		}

		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Printf("[scheduler] invalid timezone %s, fallback to UTC", tz)
			o.Location = time.UTC
			return
		}
		o.Location = loc
	}
}
