package ratelimit

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Option configures the rate limiter.
type Option func(*options)

type options struct {
	rps      rate.Limit
	burst    int
	timeout  time.Duration
	onReject func(*gin.Context, error)
}

const (
	defaultRPS     = 10
	defaultBurst   = 20
	defaultTimeout = 100 * time.Millisecond
)

func defaultOptions() *options {
	return &options{
		rps:     defaultRPS,
		burst:   defaultBurst,
		timeout: defaultTimeout,
	}
}

// WithRPS sets the rate limit (requests per second). Must be > 0.
func WithRPS(rps int) Option {
	return func(o *options) {
		if rps > 0 {
			o.rps = rate.Limit(rps)
		}
	}
}

// WithBurst sets the burst size (max tokens). Must be >= 1; if not, 1 is used.
func WithBurst(burst int) Option {
	return func(o *options) {
		if burst >= 1 {
			o.burst = burst
		}
	}
}

// WithTimeout sets the max wait duration for WaitMiddleware. Must be > 0.
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.timeout = d
		}
	}
}

// WithOnReject sets a custom handler when the request is rejected (rate exceeded or wait timeout).
// If nil, the default JSON response (429 or 408) is used.
func WithOnReject(fn func(*gin.Context, error)) Option {
	return func(o *options) {
		o.onReject = fn
	}
}
