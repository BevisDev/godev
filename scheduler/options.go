package scheduler

import (
	"log"
	"time"
)

type Option func(*Options)

type Options struct {
	Location    *time.Location
	WithSeconds bool
}

func defaultOptions() *Options {
	return &Options{
		Location:    time.UTC,
		WithSeconds: false,
	}
}

func WithSeconds() Option {
	return func(o *Options) {
		o.WithSeconds = true
	}
}

func WithLocation(loc *time.Location) Option {
	return func(o *Options) {
		if loc != nil {
			o.Location = loc
		}
	}
}

func WithTimezone(tz string) Option {
	return func(o *Options) {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Printf("[scheduler] invalid timezone %s, fallback to UTC", tz)
			o.Location = time.UTC
			return
		}
		o.Location = loc
	}
}
