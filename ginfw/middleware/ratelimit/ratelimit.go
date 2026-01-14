package ratelimit

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func New(fs ...OptionFunc) gin.HandlerFunc {
	o := defaultOptions()
	for _, opt := range fs {
		if opt != nil {
			opt(o)
		}
	}

	limiter := rate.NewLimiter(o.rps, o.burst)
	switch o.mode {
	case WaitMode:
		return rateLimitWait(limiter, o)
	default:
		return rateLimitAllow(limiter)
	}
}

func rateLimitAllow(limiter *rate.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": "TOO_MANY_REQUESTS",
			})
			return
		}

		c.Next()
	}
}

func rateLimitWait(limiter *rate.Limiter, o *options) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), o.timeout)
		defer cancel()

		if err := limiter.Wait(ctx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
					"code": "REQUEST_TIMEOUT",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": "TOO_MANY_REQUESTS",
			})
			return
		}

		c.Next()
	}
}
