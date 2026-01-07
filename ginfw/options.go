package ginfw

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Options struct {
	Profile string
	Port    string
	Proxies []string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	Setup func(r *gin.Engine)

	Shutdown func()
}
