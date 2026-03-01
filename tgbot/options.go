package tgbot

import "time"

type Option func(*options)

type options struct {
	sessionDuration time.Duration
}

func withDefaults() *options {
	return &options{
		sessionDuration: 1 * time.Hour,
	}
}

func WithSessionDuration(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.sessionDuration = d
		}
	}
}
