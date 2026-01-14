package timeout

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type OptionFunc func(*options)

type options struct {
	duration time.Duration
	response func(*gin.Context)
}

func WithTimeout(d time.Duration) OptionFunc {
	return func(o *options) {
		if d > 0 {
			o.duration = d
		}
	}
}

func WithResponse(fn func(*gin.Context)) OptionFunc {
	return func(o *options) {
		if fn != nil {
			o.response = fn
		}
	}
}

func withDefaults() *options {
	return &options{
		duration: 1 * time.Minute,
		response: func(c *gin.Context) {
			c.AbortWithStatus(http.StatusGatewayTimeout)
		},
	}
}
