package ginfw

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type Options struct {
	IsProduction bool
	Port         string
	Proxies      []string

	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration

	Setup    func(r *gin.Engine)
	Shutdown func(ctx context.Context) error
}

func (o *Options) withDefault() {
	if o.Port == "" {
		o.Port = "8080"
	}

	if o.ShutdownTimeout == 0 {
		o.ShutdownTimeout = 30 * time.Second
	}
	if o.ReadTimeout == 0 {
		o.ReadTimeout = 10 * time.Second
	}
	if o.WriteTimeout == 0 {
		o.WriteTimeout = 10 * time.Second
	}
	if o.IdleTimeout == 0 {
		o.IdleTimeout = 60 * time.Second
	}
}
