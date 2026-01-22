# HTTP Server Package (`ginfw/server`)

The `server` package provides a production-ready HTTP server built on top of [Gin](https://github.com/gin-gonic/gin) framework with graceful shutdown, lifecycle management, and flexible configuration.

---

## Features

- ✅ **Gin Framework Integration**: Built on top of Gin for high performance
- ✅ **Graceful Shutdown**: Automatic shutdown handling with configurable timeout
- ✅ **Lifecycle Management**: Separate `Start()` and `Stop()` methods for fine-grained control
- ✅ **Production Ready**: Automatic mode switching (debug/production)
- ✅ **Custom Recovery**: Configurable panic recovery middleware
- ✅ **Trusted Proxies**: Support for reverse proxy configurations
- ✅ **Signal Handling**: Automatic SIGINT/SIGTERM handling in `Run()` method

---

## Structure

### `Config`

Configuration struct for the HTTP server:

| Field              | Type                          | Description                                                      |
|-------------------|-------------------------------|------------------------------------------------------------------|
| `IsProduction`    | `bool`                        | Enable production mode (release mode, minimal logging)          |
| `Port`            | `string`                      | TCP port to listen on (e.g., `"8080"`)                         |
| `Proxies`         | `[]string`                    | List of trusted proxy IPs or CIDRs                              |
| `ShutdownTimeout` | `time.Duration`               | Maximum duration for graceful shutdown (default: 15s)           |
| `ReadHeaderTimeout` | `time.Duration`             | Timeout for reading request headers (default: 5s)               |
| `ReadTimeout`     | `time.Duration`               | Timeout for reading entire request (default: 10s)               |
| `WriteTimeout`    | `time.Duration`               | Timeout for writing response (default: 15s)                     |
| `IdleTimeout`     | `time.Duration`               | Timeout for idle connections (default: 60s)                     |
| `Setup`           | `func(r *gin.Engine)`         | Hook to configure routes and middlewares                        |
| `Shutdown`        | `func(ctx context.Context) error` | Hook for cleanup during shutdown                               |
| `Recovery`        | `func(c *gin.Context, err any)` | Custom panic recovery handler                                  |

### `HTTPApp`

Main server struct that manages the HTTP server lifecycle.

| Method                    | Description                                    |
|---------------------------|------------------------------------------------|
| `New(cfg *Config) *HTTPApp` | Create a new HTTP server instance             |
| `Start() error`           | Start the server in a goroutine (non-blocking) |
| `Stop(ctx context.Context) error` | Gracefully stop the server                  |
| `Run(ctx context.Context) error` | Start and wait for shutdown signals         |

---

## Quick Start

### Basic Usage

```go
package main

import (
	"context"
	"log"

	"github.com/BevisDev/godev/ginfw/server"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	app := server.New(&server.Config{
		Port:         "8080",
		IsProduction: false,
		Setup: func(r *gin.Engine) {
			r.GET("/health", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})
		},
	})

	// Start and wait for shutdown
	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

### Manual Lifecycle Control

```go
app := server.New(&server.Config{
	Port: "8080",
	Setup: func(r *gin.Engine) {
		r.GET("/api/users", getUsersHandler)
	},
})

// Start server (non-blocking)
if err := app.Start(); err != nil {
	log.Fatal(err)
}

// ... do other work ...

// Stop gracefully
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := app.Stop(ctx); err != nil {
	log.Printf("Shutdown error: %v", err)
}
```

### With Shutdown Hook

```go
app := server.New(&server.Config{
	Port: "8080",
	Setup: func(r *gin.Engine) {
		r.GET("/health", healthHandler)
	},
	Shutdown: func(ctx context.Context) error {
		// Cleanup resources (close DB, Redis, etc.)
		log.Println("Cleaning up resources...")
		return nil
	},
})

app.Run(context.Background())
```

### Production Configuration

```go
app := server.New(&server.Config{
	Port:         "8080",
	IsProduction: true,
	Proxies:      []string{"127.0.0.1", "10.0.0.0/8"},
	ShutdownTimeout: 30 * time.Second,
	Setup: func(r *gin.Engine) {
		// Setup routes
		api := r.Group("/api")
		api.GET("/users", getUsersHandler)
		api.POST("/users", createUserHandler)
	},
	Recovery: func(c *gin.Context, err any) {
		// Custom recovery handler
		log.Printf("Panic recovered: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
	},
})
```

---

## Integration with Framework

The server package is designed to work seamlessly with the `framework` package:

```go
import (
	"github.com/BevisDev/godev/framework"
	"github.com/BevisDev/godev/ginfw/server"
)

bootstrap := framework.New(
	framework.WithServer(&server.Config{
		Port: "8080",
		Setup: func(r *gin.Engine) {
			r.GET("/health", healthHandler)
		},
	}),
)

bootstrap.Run(ctx)
```

---

## Best Practices

1. **Always use Setup hook**: Configure routes and middlewares in the `Setup` function
2. **Handle shutdown gracefully**: Use `Shutdown` hook to cleanup resources
3. **Set appropriate timeouts**: Configure timeouts based on your application needs
4. **Use production mode**: Set `IsProduction: true` in production environments
5. **Configure trusted proxies**: Set `Proxies` when behind a reverse proxy

---

## API Reference

### `New(cfg *Config) *HTTPApp`

Creates a new HTTP server instance. The server is not started until `Start()` or `Run()` is called.

### `Start() error`

Starts the HTTP server in a goroutine. Returns immediately after starting. Use this for manual lifecycle control.

### `Stop(ctx context.Context) error`

Gracefully stops the HTTP server. It:
1. Calls the `Shutdown` hook if configured
2. Stops accepting new connections
3. Waits for ongoing requests to complete (up to `ShutdownTimeout`)
4. Closes the server

### `Run(ctx context.Context) error`

Convenience method that:
1. Calls `Start()` to start the server
2. Waits for shutdown signal (SIGINT/SIGTERM) or context cancellation
3. Calls `Stop()` for graceful shutdown

---

## Notes

- The server uses Gin's default recovery middleware in production mode unless a custom `Recovery` is provided
- In debug mode, Gin's default logger middleware is enabled
- The server automatically handles `http.ErrServerClosed` and doesn't treat it as an error
- All timeouts have sensible defaults but should be adjusted based on your application needs
