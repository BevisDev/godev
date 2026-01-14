package timeout

import (
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func Timeout(fs ...OptionFunc) gin.HandlerFunc {
	o := withDefaults()
	for _, f := range fs {
		f(o)
	}

	return timeout.New(
		timeout.WithTimeout(o.duration),
		timeout.WithResponse(o.response),
	)
}
