package server

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Config defines the configuration for running a Gin HTTP server.
type Config struct {
	// IsProduction indicates whether the server is running in production mode.
	// When true, Gin will run in release mode with minimal logging.
	IsProduction bool

	// Port is the TCP port the HTTP server listens on.
	// Example: "8080"
	Port string

	// Proxies defines a list of trusted proxy IPs or CIDRs.
	// If empty, no trusted proxies are configured.
	Proxies []string

	// ShutdownTimeout is the maximum duration the server waits
	// for ongoing requests to finish during graceful shutdown.
	ShutdownTimeout time.Duration

	// ReadHeaderTimeout limits the time to read HTTP request headers.
	// Zero means no timeout.
	ReadHeaderTimeout time.Duration

	// ReadTimeout is the maximum duration for reading the entire request,
	// including the body.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait
	// for the next request when keep-alives are enabled.
	IdleTimeout time.Duration

	// Setup is an optional hook to configure the Gin engine before the server starts.
	//
	// This is the main composition point for the HTTP layer.
	// Typical use cases include:
	//   - Registering middlewares
	//   - Defining routes and route groups
	//   - Mounting controllers
	//   - Configuring static files, metrics, or Swagger
	//
	// Setup is called exactly once during server startup.
	Setup func(r *gin.Engine)

	// Shutdown is an optional hook invoked during graceful shutdown.
	//
	// It is executed before the HTTP server shuts down and should be used
	// to release resources such as database connections, message consumers,
	// or background workers.
	//
	// The provided context is canceled when the shutdown timeout is reached.
	Shutdown func(ctx context.Context) error

	// Recovery is an optional custom panic recovery middleware.
	Recovery func(c *gin.Context, err any)
}

func (c *Config) clone() *Config {
	clone := &Config{
		IsProduction:      c.IsProduction,
		Port:              c.Port,
		Proxies:           c.Proxies,
		ShutdownTimeout:   c.ShutdownTimeout,
		ReadHeaderTimeout: c.ReadTimeout,
		ReadTimeout:       c.ReadTimeout,
		WriteTimeout:      c.WriteTimeout,
		IdleTimeout:       c.IdleTimeout,
		Setup:             c.Setup,
		Shutdown:          c.Shutdown,
		Recovery:          c.Recovery,
	}
	if clone.Port == "" {
		clone.Port = "8080"
	}
	if clone.ShutdownTimeout <= 0 {
		clone.ShutdownTimeout = 15 * time.Second
	}
	if clone.ReadHeaderTimeout <= 0 {
		clone.ReadHeaderTimeout = 5 * time.Second
	}
	if clone.ReadTimeout <= 0 {
		clone.ReadTimeout = 10 * time.Second
	}
	if clone.WriteTimeout <= 0 {
		clone.WriteTimeout = 15 * time.Second
	}
	if clone.IdleTimeout <= 0 {
		clone.IdleTimeout = 60 * time.Second
	}
	return clone
}
