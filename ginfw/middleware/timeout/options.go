package timeout

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Option func(*options)

type options struct {
	requestTimeout time.Duration
	onTimeout      func(*gin.Context)
}

func WithTimeout(duration time.Duration) Option {
	return func(o *options) {
		if duration > 0 {
			o.requestTimeout = duration
		}
	}
}

func WithResponse(onTimeout func(*gin.Context)) Option {
	return func(o *options) {
		if onTimeout != nil {
			o.onTimeout = onTimeout
		}
	}
}

func defaultOptions() *options {
	return &options{
		requestTimeout: 1 * time.Minute,
		onTimeout: func(c *gin.Context) {
			c.AbortWithStatus(http.StatusGatewayTimeout)
		},
	}
}
