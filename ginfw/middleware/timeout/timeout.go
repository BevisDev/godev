package timeout

import (
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

type Timeout struct {
	*options
}

func New(fs ...OptionFunc) *Timeout {
	o := withDefaults()
	for _, f := range fs {
		if f != nil {
			f(o)
		}
	}

	return &Timeout{
		options: o,
	}
}

func (t *Timeout) Handler() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(t.duration),
		timeout.WithResponse(t.response),
	)
}
