package ratelimit

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	codeTooManyRequests = "TOO_MANY_REQUESTS"
	codeRequestTimeout  = "REQUEST_TIMEOUT"
)

type RateLimit struct {
	*options
	limiter *rate.Limiter
}

func New(fs ...Option) *RateLimit {
	o := defaultOptions()
	for _, opt := range fs {
		opt(o)
	}

	return &RateLimit{
		options: o,
		limiter: rate.NewLimiter(o.rps, o.burst),
	}
}

func (r *RateLimit) AllowHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.limiter.Allow() {
			r.reject(c, errors.New("rate limit exceeded"))
			return
		}
		c.Next()
	}
}

func (r *RateLimit) WaitHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), r.timeout)
		defer cancel()

		if err := r.limiter.Wait(ctx); err != nil {
			r.reject(c, err)
			return
		}

		c.Next()
	}
}

func (r *RateLimit) reject(c *gin.Context, err error) {
	if r.onReject != nil {
		r.onReject(c, err)
		return
	}

	// Default error handling
	if errors.Is(err, context.DeadlineExceeded) {
		c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
			"code": codeRequestTimeout,
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"code": codeTooManyRequests,
	})
}
