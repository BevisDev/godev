# Timeout Middleware (`ginfw/middleware/timeout`)

The `timeout` middleware provides request timeout functionality for Gin applications. It automatically cancels requests that exceed a specified duration and returns a configurable timeout response.

---

## Features

- ✅ **Request Timeout**: Automatically cancel requests that exceed timeout
- ✅ **Configurable Duration**: Set custom timeout per middleware instance
- ✅ **Custom Response**: Configurable timeout response handler
- ✅ **Context Cancellation**: Properly cancels request context on timeout
- ✅ **Built on gin-contrib/timeout**: Uses battle-tested timeout middleware

---

## Structure

### `Timeout`

Main middleware struct that handles request timeouts.

| Method | Description |
|--------|-------------|
| `New(opts ...OptionFunc) *Timeout` | Create a new timeout middleware instance |
| `Handler() gin.HandlerFunc` | Returns the Gin middleware handler function |

### Options

| Option | Description |
|--------|-------------|
| `WithTimeout(duration time.Duration)` | Set request timeout (default: 5s) |
| `WithResponse(fn func(c *gin.Context))` | Custom timeout response handler |

---

## Quick Start

### Basic Usage

```go
package main

import (
	"time"

	"github.com/BevisDev/godev/ginfw/middleware/timeout"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Add timeout middleware: 10 seconds
	r.Use(timeout.New(
		timeout.WithTimeout(10 * time.Second),
	).Handler())

	r.GET("/api/slow", slowHandler)
	r.Run(":8080")
}
```

### Custom Timeout Response

```go
r.Use(timeout.New(
	timeout.WithTimeout(5 * time.Second),
	timeout.WithResponse(func(c *gin.Context) {
		// Custom timeout response
		c.JSON(408, gin.H{
			"error": "Request timeout",
			"message": "The request took too long to process",
		})
	}),
).Handler())
```

### Per-Route Timeout

```go
// Global timeout: 10 seconds
r.Use(timeout.New(
	timeout.WithTimeout(10 * time.Second),
).Handler())

// Specific route with shorter timeout
api := r.Group("/api")
api.Use(timeout.New(
	timeout.WithTimeout(3 * time.Second),
).Handler())
api.GET("/fast", fastHandler)
```

---

## Default Behavior

When a request exceeds the timeout:

1. The request context is cancelled
2. The default response handler is called (if not customized)
3. Default response: `408 Request Timeout` with JSON:
   ```json
   {
     "message": "Request timeout"
   }
   ```

---

## Custom Response Handler

You can customize the timeout response:

```go
r.Use(timeout.New(
	timeout.WithTimeout(5 * time.Second),
	timeout.WithResponse(func(c *gin.Context) {
		// Log timeout event
		logger.Warn("timeout", "Request timeout", 
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)
		
		// Custom response
		c.Header("X-Timeout-Duration", "5s")
		c.JSON(408, gin.H{
			"error": "timeout",
			"message": "Request exceeded timeout",
			"timeout": "5s",
		})
	}),
).Handler())
```

---

## Best Practices

1. **Set appropriate timeouts**: Base on your handler's expected execution time
2. **Use shorter timeouts for public APIs**: Protect against slow clients
3. **Use longer timeouts for background jobs**: Allow time for processing
4. **Monitor timeout events**: Log timeouts to identify performance issues
5. **Handle context cancellation**: Check `ctx.Done()` in long-running handlers
6. **Use per-route timeouts**: Different endpoints may need different timeouts

---

## Handler Implementation

When using timeout middleware, handlers should respect context cancellation:

```go
func slowHandler(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return // Request already timed out
	default:
	}
	
	// Long-running operation
	result := processData(ctx)
	
	// Check context again before responding
	select {
	case <-ctx.Done():
		return // Timeout occurred during processing
	default:
		c.JSON(200, result)
	}
}
```

---

## Integration with Framework

```go
import (
	"github.com/BevisDev/godev/framework"
	"github.com/BevisDev/godev/ginfw/middleware/timeout"
	"github.com/BevisDev/godev/ginfw/server"
)

bootstrap := framework.New(
	framework.WithServer(&server.Config{
		Port: "8080",
		Setup: func(r *gin.Engine) {
			// Add timeout middleware
			r.Use(timeout.New(
				timeout.WithTimeout(10 * time.Second),
			).Handler())
			
			r.GET("/api/data", getDataHandler)
		},
	}),
)
```

---

## Notes

- The timeout applies to the entire request processing time
- Context cancellation happens automatically when timeout is reached
- Handlers should check `ctx.Done()` to stop processing early
- The middleware uses `gin-contrib/timeout` internally
- Default timeout is 5 seconds if not specified
- Timeout response is sent immediately when timeout is reached
