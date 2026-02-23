package ratelimit

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Response codes returned in JSON when the request is rejected.
const (
	CodeTooManyRequests = "TOO_MANY_REQUESTS"
	CodeRequestTimeout  = "REQUEST_TIMEOUT"
)

// RateLimiter applies rate limiting to HTTP requests using golang.org/x/time/rate.
type RateLimiter struct {
	opts    *options
	limiter *rate.Limiter
}

// RateLimit is an alias for RateLimiter for backward compatibility.
type RateLimit = RateLimiter

// New returns a new RateLimiter with the given options.
func New(opts ...Option) *RateLimiter {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	burst := max(o.burst, 1)
	return &RateLimiter{
		opts:    o,
		limiter: rate.NewLimiter(o.rps, burst),
	}
}

// Middleware returns a Gin middleware that allows requests up to the rate limit
// and rejects with 429 (Too Many Requests) when exceeded. Non-blocking.
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return r.AllowHandler()
}

// AllowHandler returns a Gin middleware that allows requests when a token is
// available and aborts with 429 when the rate is exceeded (non-blocking).
func (r *RateLimiter) AllowHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.limiter.Allow() {
			r.reject(c, errors.New("rate limit exceeded"))
			return
		}
		c.Next()
	}
}

// WaitHandler returns a Gin middleware that waits up to the configured timeout
// for a token; if the wait times out or the context is cancelled, it aborts with
// 408 (Request Timeout) or 429 respectively.
func (r *RateLimiter) WaitHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), r.opts.timeout)
		defer cancel()

		if err := r.limiter.Wait(ctx); err != nil {
			r.reject(c, err)
			return
		}
		c.Next()
	}
}

func (r *RateLimiter) reject(c *gin.Context, err error) {
	if r.opts.onReject != nil {
		r.opts.onReject(c, err)
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
			"code":    CodeRequestTimeout,
			"message": "rate limit wait timeout",
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"code":    CodeTooManyRequests,
		"message": "rate limit exceeded",
	})
}
