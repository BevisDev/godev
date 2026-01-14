package ratelimit

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type OptionFunc func(*options)

type options struct {
	rps      rate.Limit
	burst    int
	timeout  time.Duration
	mode     Mode
	onReject func(c *gin.Context, err error)
}

func defaultOptions() *options {
	return &options{
		rps:     10,
		burst:   20,
		timeout: 100 * time.Millisecond,
		mode:    AllowMode,
	}
}

func WithRPS(rps int) OptionFunc {
	return func(o *options) {
		if rps > 0 {
			o.rps = rate.Limit(rps)
		}
	}
}

func WithBurst(burst int) OptionFunc {
	return func(o *options) {
		if burst > 0 {
			o.burst = burst
		}
	}
}

func WithTimeout(d time.Duration) OptionFunc {
	return func(o *options) {
		if d > 0 {
			o.timeout = d
		}
	}
}

func WithOnReject(fn func(c *gin.Context, err error)) OptionFunc {
	return func(o *options) {
		o.onReject = fn
	}
}

func WithMode(Mode Mode) OptionFunc {
	return func(o *options) {
		o.mode = Mode
	}
}
