# HTTP Logger Middleware (`ginfw/middleware/logger`)

The `logger` middleware provides comprehensive HTTP request and response logging for Gin applications. It supports both structured logging (via `logx`) and console logging, with flexible configuration options.

---

## Features

- ✅ **Request/Response Logging**: Logs all HTTP requests and responses with detailed information
- ✅ **Request ID (RID)**: Automatically generates and tracks unique request IDs
- ✅ **Dual Logging Modes**: Supports both structured logging (`logx`) and console logging
- ✅ **Body Filtering**: Skip logging request/response bodies based on content type
- ✅ **Header Control**: Optional header logging
- ✅ **Context Integration**: Automatically attaches RID to request context

---

## Structure

### `HttpLogger`

Main middleware struct that handles HTTP logging.

| Method | Description |
|--------|-------------|
| `New(opts ...OptionFunc) *HttpLogger` | Create a new logger middleware instance |
| `Handler() gin.HandlerFunc` | Returns the Gin middleware handler function |

### Options

| Option | Description |
|--------|-------------|
| `WithLogger(logger logx.Logger)` | Enable structured logging with logx |
| `WithSkipHeader()` | Skip logging HTTP headers |
| `WithSkipBodyByPaths(...string)` | Skip body logging for specific paths |
| `WithSkipBodyByContentTypes(...string)` | Skip body logging for specific content types |
| `WithSkipDefaultContentTypeCheck()` | Disable default content-type based filtering |

---

## Quick Start

### Basic Usage (Console Logging)

```go
package main

import (
	"github.com/BevisDev/godev/ginfw/middleware/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Add logger middleware
	r.Use(logger.New().Handler())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}
```

### With Structured Logging

```go
import (
	"github.com/BevisDev/godev/ginfw/middleware/logger"
	"github.com/BevisDev/godev/logx"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	appLogger := logx.New(&logx.Config{
		IsProduction: true,
		DirName:      "./logs",
		Filename:     "app.log",
	})

	r := gin.Default()

	// Add logger middleware with structured logging
	r.Use(logger.New(
		logger.WithLogger(appLogger),
	).Handler())

	r.GET("/api/users", getUsersHandler)
	r.Run(":8080")
}
```

### Advanced Configuration

```go
r.Use(logger.New(
	logger.WithLogger(appLogger),
	logger.WithSkipHeader(), // Don't log headers
	logger.WithSkipBodyByPaths("/health", "/metrics"), // Skip body for these paths
	logger.WithSkipBodyByContentTypes("image/*", "video/*"), // Skip body for media
).Handler())
```

---

## Logged Information

### Request Logging

The middleware logs the following request information:

- **RID**: Unique request identifier (UUID)
- **URL**: Full request URL
- **Method**: HTTP method (GET, POST, etc.)
- **Time**: Request timestamp
- **Query**: Query parameters
- **Header**: HTTP headers (optional)
- **Body**: Request body (filtered by content type)

### Response Logging

The middleware logs the following response information:

- **RID**: Request identifier (matches request)
- **Status**: HTTP status code
- **Duration**: Request processing time
- **Header**: Response headers (optional)
- **Body**: Response body (filtered by content type)

---

## Example Output

### Console Mode

```
========== REQUEST INFO ==========
RID: 550e8400-e29b-71d4-a716-446655440000
URL: /api/users?page=1
Method: GET
Time: 2024-01-15 10:30:45.123
Query: page=1
Body: 
==================================

========== RESPONSE INFO ==========
RID: 550e8400-e29b-71d4-a716-446655440000
Status: 200
Duration: 45.2ms
Body: {"users":[...]}
==================================
```

### Structured Logging Mode

When using `logx`, logs are written in JSON format:

```json
{
  "level": "info",
  "rid": "550e8400-e29b-71d4-a716-446655440000",
  "url": "/api/users",
  "method": "GET",
  "time": "2024-01-15T10:30:45.123Z",
  "query": "page=1",
  "body": ""
}
```

---

## Accessing Request ID

The middleware automatically attaches the RID to the request context. You can retrieve it in your handlers:

```go
import (
	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/gin-gonic/gin"
)

func handler(c *gin.Context) {
	// Get RID from context
	rid := utils.GetRID(c.Request.Context())
	
	// Use RID in your logic
	c.JSON(200, gin.H{
		"rid": rid,
		"data": "response",
	})
}
```

---

## Best Practices

1. **Use structured logging in production**: Enable `logx` logger for better log management
2. **Skip sensitive data**: Use `WithSkipBodyByPaths` for endpoints that handle sensitive information
3. **Filter media files**: Use `WithSkipBodyByContentTypes` to avoid logging large binary files
4. **Skip headers when not needed**: Use `WithSkipHeader()` to reduce log size
5. **Use RID for tracing**: Access RID from context for distributed tracing

---

## Integration with Framework

The logger middleware works seamlessly with the framework:

```go
import (
	"github.com/BevisDev/godev/framework"
	"github.com/BevisDev/godev/ginfw/middleware/logger"
	"github.com/BevisDev/godev/ginfw/server"
)

bootstrap := framework.New(
	framework.WithServer(&server.Config{
		Port: "8080",
		Setup: func(r *gin.Engine) {
			// Add logger middleware
			r.Use(logger.New(
				logger.WithLogger(bootstrap.Logger),
			).Handler())
			
			r.GET("/health", healthHandler)
		},
	}),
)
```

---

## Notes

- The middleware automatically generates a UUID for each request
- Request body is read and restored, so handlers can still read it
- Response body is captured using a custom response writer wrapper
- Content type filtering uses default rules (images, videos, etc.) unless disabled
- RID is available in the request context for use in handlers
