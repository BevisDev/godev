package timeout

import (
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

type Timeout struct {
	*options
}

func New(opts ...Option) *Timeout {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Timeout{
		options: o,
	}
}

func (t *Timeout) Handler() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(t.requestTimeout),
		timeout.WithResponse(t.onTimeout),
	)
}
