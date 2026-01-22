# Rate Limit Middleware (`ginfw/middleware/ratelimit`)

The `ratelimit` middleware provides rate limiting functionality for Gin applications using the token bucket algorithm. It supports two modes: **Allow** (immediate rejection) and **Wait** (queue requests).

---

## Features

- ✅ **Token Bucket Algorithm**: Uses `golang.org/x/time/rate` for efficient rate limiting
- ✅ **Two Modes**: Allow mode (immediate rejection) and Wait mode (queue requests)
- ✅ **Configurable RPS**: Set requests per second and burst capacity
- ✅ **Custom Error Handling**: Configurable rejection handler
- ✅ **Timeout Support**: Configurable timeout for Wait mode
- ✅ **Thread-Safe**: Safe for concurrent use

---

## Structure

### `RateLimit`

Main middleware struct that manages rate limiting.

| Method | Description |
|--------|-------------|
| `New(opts ...OptionFunc) *RateLimit` | Create a new rate limiter instance |
| `AllowHandler() gin.HandlerFunc` | Returns handler that immediately rejects when limit exceeded |
| `WaitHandler() gin.HandlerFunc` | Returns handler that queues requests when limit exceeded |

### Options

| Option | Description |
|--------|-------------|
| `WithRPS(rps float64)` | Set requests per second (default: 10) |
| `WithBurst(burst int)` | Set burst capacity (default: 10) |
| `WithMode(mode Mode)` | Set rate limit mode: `AllowMode` or `WaitMode` |
| `WithTimeout(timeout time.Duration)` | Set timeout for Wait mode (default: 5s) |
| `WithOnReject(fn func(c *gin.Context, err error))` | Custom rejection handler |

---

## Quick Start

### Allow Mode (Immediate Rejection)

```go
package main

import (
	"github.com/BevisDev/godev/ginfw/middleware/ratelimit"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Add rate limiter: 10 requests per second, burst of 20
	r.Use(ratelimit.New(
		ratelimit.WithRPS(10),
		ratelimit.WithBurst(20),
		ratelimit.WithMode(ratelimit.AllowMode),
	).AllowHandler())

	r.GET("/api/data", getDataHandler)
	r.Run(":8080")
}
```

### Wait Mode (Queue Requests)

```go
r.Use(ratelimit.New(
	ratelimit.WithRPS(10),
	ratelimit.WithBurst(20),
	ratelimit.WithMode(ratelimit.WaitMode),
	ratelimit.WithTimeout(5 * time.Second), // Wait up to 5 seconds
).WaitHandler())
```

### Custom Rejection Handler

```go
r.Use(ratelimit.New(
	ratelimit.WithRPS(10),
	ratelimit.WithBurst(20),
	ratelimit.WithOnReject(func(c *gin.Context, err error) {
		// Custom error response
		c.JSON(429, gin.H{
			"error": "Rate limit exceeded",
			"retry_after": 1,
		})
	}),
).AllowHandler())
```

---

## Rate Limit Modes

### Allow Mode

In Allow mode, requests that exceed the rate limit are immediately rejected with a `429 Too Many Requests` status.

**Use cases:**
- API endpoints that should fail fast
- Public APIs with strict rate limits
- Endpoints where queuing doesn't make sense

**Example:**
```go
r.Use(ratelimit.New(
	ratelimit.WithRPS(100),      // 100 requests per second
	ratelimit.WithBurst(200),     // Allow bursts up to 200
	ratelimit.WithMode(ratelimit.AllowMode),
).AllowHandler())
```

### Wait Mode

In Wait mode, requests that exceed the rate limit are queued and wait until tokens become available, up to a configurable timeout.

**Use cases:**
- Background processing endpoints
- Internal APIs where queuing is acceptable
- Endpoints that benefit from request smoothing

**Example:**
```go
r.Use(ratelimit.New(
	ratelimit.WithRPS(50),                    // 50 requests per second
	ratelimit.WithBurst(100),                  // Allow bursts up to 100
	ratelimit.WithMode(ratelimit.WaitMode),
	ratelimit.WithTimeout(10 * time.Second),  // Wait up to 10 seconds
).WaitHandler())
```

---

## Error Responses

### Default Error Handling

When rate limit is exceeded:

- **Allow Mode**: Returns `429 Too Many Requests` with JSON:
  ```json
  {
    "code": "TOO_MANY_REQUESTS"
  }
  ```

- **Wait Mode (timeout)**: Returns `408 Request Timeout` with JSON:
  ```json
  {
    "code": "REQUEST_TIMEOUT"
  }
  ```

### Custom Error Handling

You can provide a custom rejection handler:

```go
r.Use(ratelimit.New(
	ratelimit.WithRPS(10),
	ratelimit.WithOnReject(func(c *gin.Context, err error) {
		// Log the rate limit event
		logger.Warn("rate_limit", "Rate limit exceeded", "ip", c.ClientIP())
		
		// Custom response
		c.Header("Retry-After", "60")
		c.JSON(429, gin.H{
			"error": "Too many requests",
			"message": "Please try again later",
			"retry_after": 60,
		})
	}),
).AllowHandler())
```

---

## Token Bucket Algorithm

The rate limiter uses a token bucket algorithm:

- **RPS (Rate)**: The rate at which tokens are added to the bucket
- **Burst**: The maximum number of tokens the bucket can hold
- **Tokens**: Each request consumes one token

**Example:**
- RPS: 10 (10 tokens per second)
- Burst: 20 (bucket can hold 20 tokens)

This means:
- Steady state: 10 requests per second
- Burst: Up to 20 requests can be processed immediately
- After burst: Requests are limited to 10 per second

---

## Best Practices

1. **Set appropriate RPS**: Base on your server capacity and expected load
2. **Configure burst**: Allow for traffic spikes while maintaining overall rate
3. **Use Allow mode for public APIs**: Fail fast to protect your backend
4. **Use Wait mode for internal APIs**: Smooth out traffic spikes
5. **Set reasonable timeouts**: Don't let requests wait indefinitely
6. **Monitor rate limit hits**: Use custom rejection handler to log events
7. **Consider per-IP limiting**: Combine with IP-based middleware for more granular control

---

## Integration with Framework

```go
import (
	"github.com/BevisDev/godev/framework"
	"github.com/BevisDev/godev/ginfw/middleware/ratelimit"
	"github.com/BevisDev/godev/ginfw/server"
)

bootstrap := framework.New(
	framework.WithServer(&server.Config{
		Port: "8080",
		Setup: func(r *gin.Engine) {
			// Add rate limiter
			r.Use(ratelimit.New(
				ratelimit.WithRPS(100),
				ratelimit.WithBurst(200),
				ratelimit.WithMode(ratelimit.AllowMode),
			).AllowHandler())
			
			r.GET("/api/data", getDataHandler)
		},
	}),
)
```

---

## Notes

- The rate limiter is thread-safe and can be used across multiple goroutines
- Token bucket state is shared across all requests using the same middleware instance
- For per-route rate limiting, create separate `RateLimit` instances
- The timeout in Wait mode applies to the wait duration, not the total request time
- Custom rejection handlers should call `c.Abort()` or return a response
