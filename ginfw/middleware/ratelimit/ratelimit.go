package ratelimit

import (
	"context"
	"errors"

	"github.com/BevisDev/godev/ginfw/response"
	"github.com/BevisDev/godev/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// RateLimiter applies rate limiting to HTTP requests using golang.org/x/time/rate.
type RateLimiter struct {
	*options
	limiter *rate.Limiter
}

// New returns a new RateLimiter with the given options.
func New(opts ...Option) *RateLimiter {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	burst := max(o.burst, 1)
	return &RateLimiter{
		options: o,
		limiter: rate.NewLimiter(o.rps, burst),
	}
}

// AllowHandler returns a Gin middleware that allows requests when a token is
// available and aborts with 429 when the rate is exceeded (non-blocking).
func (r *RateLimiter) AllowHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.limiter.Allow() {
			r.reject(c, ErrRateLimitExceeded)
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
		ctx, cancel := utils.NewCtxTimeout(c.Request.Context(), r.waitTimeout)
		defer cancel()

		if err := r.limiter.Wait(ctx); err != nil {
			r.reject(c, err)
			return
		}
		c.Next()
	}
}

func (r *RateLimiter) reject(c *gin.Context, err error) {
	if r.onReject != nil {
		r.onReject(c, err)
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		response.RequestTimeout(c, "", "")
		return
	}
	response.TooManyRequests(c, "", "")
}
