# 🌟 GoDev

**GoDev** is a comprehensive Go development toolkit that provides essential utilities, integrations, and framework components for building robust applications. It simplifies common development tasks and accelerates project setup with pre-built, production-ready packages.

---

## ✨ Features

- 🗄️ **Database**: Multi-database support (PostgreSQL, MySQL, SQL Server, Oracle) with query builder
- 📝 **Logging**: Structured logging with Zap, file rotation, and HTTP request/response logging
- 🔄 **Redis**: Full-featured Redis client with pub/sub, chain operations, and JSON serialization
- 🐰 **RabbitMQ**: Message queue integration with publisher/consumer patterns
- 🔐 **Keycloak**: Identity and access management integration
- 🌐 **HTTP Server**: Gin-based HTTP server with middleware support
- 📡 **REST Client**: Type-safe HTTP client with automatic JSON handling
- ⏰ **Scheduler**: Cron job scheduler with timezone support
- 🏗️ **Framework**: Application bootstrap with lifecycle management
- ⚙️ **Config**: Configuration management with Viper, environment variables, and placeholders
- 🛠️ **Utils**: Comprehensive utility functions (crypto, datetime, string, validation, etc.)

---

## 📦 Packages

### Core Infrastructure

| Package | Description | README |
|---------|-------------|--------|
| **`framework`** | Application bootstrap with lifecycle management, service initialization, and graceful shutdown | [📖 Read More](framework/README.md) |
| **`config`** | Configuration management with file loading, environment variables, and placeholder expansion | [📖 Read More](config/README.md) |
| **`logx`** | Structured logging with Zap, file rotation, and HTTP logging | [📖 Read More](logx/README.md) |

### Data & Storage

| Package | Description | README |
|---------|-------------|--------|
| **`database`** | Multi-database abstraction with query builder, transactions, and bulk operations | [📖 Read More](database/README.md) |
| **`redis`** | Redis client with chain operations, pub/sub, and JSON serialization | [📖 Read More](redis/README.md) |
| **`rabbitmq`** | RabbitMQ integration with publisher/consumer patterns | [📖 Read More](rabbitmq/README.md) |
| **`migration`** | Database migration utilities | [📖 Read More](migration/README.md) |

### HTTP & Networking

| Package | Description | README |
|---------|-------------|--------|
| **`ginfw/server`** | Gin HTTP server with graceful shutdown and lifecycle hooks | [📖 Read More](ginfw/server/README.md) |
| **`ginfw/middleware/logger`** | HTTP request/response logging middleware | [📖 Read More](ginfw/middleware/logger/README.md) |
| **`ginfw/middleware/ratelimit`** | Rate limiting middleware with Allow/Wait modes | [📖 Read More](ginfw/middleware/ratelimit/README.md) |
| **`ginfw/middleware/timeout`** | Request timeout middleware | [📖 Read More](ginfw/middleware/timeout/README.md) |
| **`rest`** | Type-safe REST client with automatic JSON handling | [📖 Read More](rest/README.md) |

### Services & Integration

| Package | Description | README |
|---------|-------------|--------|
| **`keycloak`** | Keycloak identity and access management client | [📖 Read More](keycloak/README.md) |
| **`scheduler`** | Cron job scheduler with timezone support and graceful shutdown | [📖 Read More](scheduler/README.md) |

### Utilities

| Package | Description | README |
|---------|-------------|--------|
| **`utils`** | Comprehensive utility functions (crypto, datetime, string, validation, file, json, money, random) | [📖 Read More](utils/README.md) |
| **`consts`** | Common constants (content types, extensions, patterns) | [📖 Read More](consts/README.md) |
| **`types`** | Shared type definitions | [📖 Read More](types/README.md) |

---

## 🚀 Quick Start

### Installation

```bash
go get github.com/BevisDev/godev@latest
```

### Basic Example

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/framework"
	"github.com/BevisDev/godev/ginfw/server"
	"github.com/BevisDev/godev/logx"
	"github.com/BevisDev/godev/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	// Initialize application with framework
	bootstrap := framework.New(
		// Logger
		framework.WithLogger(&logx.Config{
			IsProduction: false,
			IsLocal:      true,
			DirName:      "./logs",
			Filename:     "app.log",
		}),

		// Database
		framework.WithDatabase(&database.Config{
			DBType:   database.Postgres,
			Host:     "localhost",
			Port:     5432,
			DBName:   "mydb",
			Username: "user",
			Password: "password",
			Timeout:  5 * time.Second,
		}),

		// Redis
		framework.WithRedis(&redis.Config{
			Host:    "localhost",
			Port:    6379,
			Timeout: 5 * time.Second,
		}),

		// HTTP Server
		framework.WithServer(&server.Config{
			Port:         "8080",
			IsProduction: false,
			Setup: func(r *gin.Engine) {
				r.GET("/health", func(c *gin.Context) {
					c.JSON(200, gin.H{"status": "ok"})
				})
			},
		}),
	)

	// Run application (init, start, graceful shutdown)
	if err := bootstrap.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

### Individual Package Usage

#### Database

```go
import "github.com/BevisDev/godev/database"

db, err := database.New(&database.Config{
	DBType:   database.Postgres,
	Host:     "localhost",
	Port:     5432,
	DBName:   "mydb",
	Username: "user",
	Password: "password",
})
if err != nil {
	log.Fatal(err)
}
defer db.Close()

// Query with builder
users, err := database.Builder[User](db).
	From("users").
	Where("age > ?", 18).
	FindAll(ctx)
```

#### Logger

```go
import "github.com/BevisDev/godev/logx"

logger := logx.New(&logx.Config{
	IsProduction: true,
	DirName:      "./logs",
	Filename:     "app.log",
	MaxSize:      100,
	MaxBackups:  7,
	MaxAge:      30,
})

logger.Info("state", "Application started")
```

#### Redis

```go
import "github.com/BevisDev/godev/redis"

cache, err := redis.New(&redis.Config{
	Host:    "localhost",
	Port:    6379,
	Timeout: 5 * time.Second,
})
if err != nil {
	log.Fatal(err)
}
defer cache.Close()

// Chain operations
err = redis.With[User](cache).
	Key("user:1").
	Value(user).
	Expire(10 * time.Second).
	Set(ctx)
```

#### REST Client

```go
import "github.com/BevisDev/godev/rest"

client := rest.New(
	rest.WithTimeout(10 * time.Second),
	rest.WithLogger(logger),
)

user, err := rest.NewRequest[*UserResponse](client).
	URL("https://api.example.com/users/:id").
	PathParams(map[string]string{"id": "1"}).
	GET(ctx)
```

---

## 📋 Requirements

- **Go**: 1.21 or higher
- **Database Drivers** (as needed):
  - PostgreSQL: `go get github.com/lib/pq`
  - MySQL: `go get github.com/go-sql-driver/mysql`
  - SQL Server: `go get github.com/denisenkom/go-mssqldb`
  - Oracle: `go get github.com/godror/godror@latest`

---

## 🏗️ Architecture

GoDev follows a modular architecture where each package is independent and can be used standalone or integrated through the `framework` package:

```
┌─────────────────────────────────────────┐
│           Framework (Bootstrap)         │
│  - Lifecycle Management                 │
│  - Service Initialization               │
│  - Graceful Shutdown                    │
└─────────────────────────────────────────┘
           │
           ├─── Database ─── Redis ─── RabbitMQ
           │
           ├─── Logger ─── REST Client ─── Scheduler
           │
           └─── HTTP Server (Gin)
                    │
                    ├─── Middleware: Logger, RateLimit, Timeout
                    └─── Routes & Handlers
```

---

## 📚 Documentation

Each package includes comprehensive documentation with examples:

- [Framework](framework/README.md) - Application bootstrap and lifecycle
- [Database](database/README.md) - Database operations and query builder
- [Logger](logx/README.md) - Structured logging
- [Redis](redis/README.md) - Redis client and operations
- [HTTP Server](ginfw/server/README.md) - Gin server setup
- [REST Client](rest/README.md) - HTTP client utilities
- [Config](config/README.md) - Configuration management
- [Scheduler](scheduler/README.md) - Cron job scheduling
- [Utils](utils/README.md) - Utility functions

---

## 🧪 Testing

All packages include comprehensive unit tests:

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./database/...
go test ./logx/...
```

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- Code follows Go best practices
- Tests are included for new features
- Documentation is updated
- All tests pass

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

GoDev is built on top of excellent open-source libraries:

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [sqlx](https://github.com/jmoiron/sqlx) - Database extensions
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Cron](https://github.com/robfig/cron) - Job scheduling

---

## 📞 Support

For questions, issues, or contributions, please open an issue on [GitHub](https://github.com/BevisDev/godev).

---

**Made with ❤️ for the Go community**
