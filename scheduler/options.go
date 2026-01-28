package scheduler

import (
	"log"
	"time"
)

type Option func(*options)

type options struct {
	location   *time.Location
	useSeconds bool
}

func defaultOptions() *options {
	return &options{
		location:   time.UTC,
		useSeconds: false,
	}
}

func WithSeconds() Option {
	return func(o *options) {
		o.useSeconds = true
	}
}

func WithLocation(loc *time.Location) Option {
	return func(o *options) {
		if loc != nil {
			o.location = loc
		}
	}
}

func WithTimezone(tz string) Option {
	return func(o *options) {
		if tz == "" {
			o.location = time.UTC
			return
		}

		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Printf("[scheduler] invalid timezone %s, fallback to UTC", tz)
			o.location = time.UTC
			return
		}
		o.location = loc
	}
}
