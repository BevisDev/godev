package tgbot

import "time"

// Option configures a TgBot at construction (e.g. WithSessionDuration).
type Option func(*options)

type options struct {
	sessionDuration time.Duration
}

func withDefaults() *options {
	return &options{
		sessionDuration: 1 * time.Hour,
	}
}

// WithSessionDuration sets how long a session stays active per chat; ignored if d <= 0.
func WithSessionDuration(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.sessionDuration = d
		}
	}
}
