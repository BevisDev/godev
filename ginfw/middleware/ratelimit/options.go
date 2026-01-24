package ratelimit

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Option func(*options)

type options struct {
	rps      rate.Limit
	burst    int
	timeout  time.Duration
	onReject func(c *gin.Context, err error)
}

func defaultOptions() *options {
	return &options{
		rps:     10,
		burst:   20,
		timeout: 100 * time.Millisecond,
	}
}

func WithRPS(rps int) Option {
	return func(o *options) {
		if rps > 0 {
			o.rps = rate.Limit(rps)
		}
	}
}

func WithBurst(burst int) Option {
	return func(o *options) {
		if burst > 0 {
			o.burst = burst
		}
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		if d > 0 {
			o.timeout = d
		}
	}
}

func WithOnReject(fn func(c *gin.Context, err error)) Option {
	return func(o *options) {
		o.onReject = fn
	}
}
