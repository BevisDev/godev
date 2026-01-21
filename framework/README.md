# Framework Bootstrap

Bootstrap framework giúp khởi tạo và quản lý lifecycle của application một cách nhanh chóng và dễ dàng.

## Features

- ✅ **Easy Initialization**: Khởi tạo tất cả services với option pattern
- ✅ **Lifecycle Management**: Quản lý init, start, stop tự động
- ✅ **Graceful Shutdown**: Tự động shutdown khi nhận signal
- ✅ **Health Checks**: Kiểm tra health của các services
- ✅ **Dependency Injection**: Dễ dàng truy cập các services
- ✅ **Lifecycle Hooks**: Before/After hooks cho init, start, stop

## Supported Services

- **Logger** (logx)
- **Database** (database)
- **Redis** (redis)
- **RabbitMQ** (rabbitmq)
- **Keycloak** (keycloak)
- **REST Client** (rest)
- **Scheduler** (scheduler)
- **Gin HTTP Server** (ginfw/server)

## Quick Start

### Basic Usage

```go
package main

import (
	"context"
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

	// Create bootstrap with all services
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
			DBType:    database.MySQL,
			Host:      "localhost",
			Port:      3306,
			DBName:    "mydb",
			Username:  "user",
			Password:  "password",
			Timeout:   5 * time.Second,
		}),

		// Redis
		framework.WithRedis(&redis.Config{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			Timeout:  5 * time.Second,
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

	// Run application
	if err := bootstrap.Run(ctx); err != nil {
		panic(err)
	}
}
```

### With Lifecycle Hooks

```go
bootstrap := framework.New(
	framework.WithLogger(&logx.Config{...}),
	framework.WithDatabase(&database.Config{...}),
)

// Before initialization
bootstrap.BeforeInit(func(ctx context.Context) error {
	log.Println("Preparing initialization...")
	return nil
})

// After initialization
bootstrap.AfterInit(func(ctx context.Context) error {
	log.Println("Running migrations...")
	// Run migrations here
	return nil
})

// Before start
bootstrap.BeforeStart(func(ctx context.Context) error {
	log.Println("Warming up cache...")
	return nil
})

// After start
bootstrap.AfterStart(func(ctx context.Context) error {
	log.Println("Application started successfully!")
	return nil
})

// Before stop
bootstrap.BeforeStop(func(ctx context.Context) error {
	log.Println("Saving state...")
	return nil
})

// After stop
bootstrap.AfterStop(func(ctx context.Context) error {
	log.Println("Cleanup completed")
	return nil
})

bootstrap.Run(ctx)
```

### Manual Lifecycle Control

```go
bootstrap := framework.New(
	framework.WithLogger(&logx.Config{...}),
	framework.WithDatabase(&database.Config{...}),
)

// Initialize
if err := bootstrap.Init(ctx); err != nil {
	panic(err)
}

// Start (non-blocking)
go func() {
	if err := bootstrap.Start(ctx); err != nil {
		log.Printf("Start error: %v", err)
	}
}()

// ... do other work ...

// Stop gracefully
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
bootstrap.Stop(shutdownCtx)
```

### With All Services

```go
bootstrap := framework.New(
	// Logger
	framework.WithLogger(&logx.Config{
		IsProduction: true,
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
	}),

	// Redis
	framework.WithRedis(&redis.Config{
		Host:    "localhost",
		Port:    6379,
		Timeout: 5 * time.Second,
	}),

	// RabbitMQ
	framework.WithRabbitMQ(&rabbitmq.Config{
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
		VHost:    "/",
	}),

	// Keycloak
	framework.WithKeycloak(&keycloak.Config{
		Host:  "localhost",
		Port:  8080,
		Realm: "myrealm",
	}),

	// REST Client
	framework.WithRestClient(
		rest.WithTimeout(10*time.Second),
		rest.WithBaseURL("https://api.example.com"),
	),

	// Scheduler
	framework.WithScheduler(
		scheduler.WithLocation(time.UTC),
		scheduler.WithSeconds(),
	),

	// HTTP Server
	framework.WithServer(&server.Config{
		Port:         "8080",
		IsProduction: true,
		Setup: func(r *gin.Engine) {
			// Setup routes
			r.GET("/health", healthHandler)
			r.GET("/api/users", usersHandler)
		},
		Shutdown: func(ctx context.Context) error {
			// Cleanup resources
			return nil
		},
	}),
)

bootstrap.Run(ctx)
```

### Accessing Services

```go
bootstrap := framework.New(...)

// After initialization, access services
logger := bootstrap.GetLogger()
db := bootstrap.GetDatabase()
redis := bootstrap.GetRedis()
rabbitmq := bootstrap.GetRabbitMQ()
keycloak := bootstrap.GetKeycloak()
rest := bootstrap.GetRest()
scheduler := bootstrap.GetScheduler()

// Use services
logger.Info("Application started")
users, _ := database.Builder[User](db).From("users").FindAll(ctx)
```

### Health Checks

```go
health := bootstrap.Health(ctx)
for service, err := range health {
	if err != nil {
		log.Printf("[health] %s: %v", service, err)
	} else {
		log.Printf("[health] %s: OK", service)
	}
}
```

### With Config File

```go
import (
	"github.com/BevisDev/godev/config"
	"github.com/BevisDev/godev/framework"
)

type AppConfig struct {
	Logger   logx.Config
	Database database.Config
	Redis    redis.Config
	Server   server.Config
}

func main() {
	// Load config from file
	cfg := config.MustLoad[AppConfig](&config.Config{
		Path:      "./configs",
		Extension: "yaml",
		Profile:   "dev",
		AutoEnv:   true,
	})

	// Create bootstrap with loaded config
	bootstrap := framework.New(
		framework.WithLogger(&cfg.Data.Logger),
		framework.WithDatabase(&cfg.Data.Database),
		framework.WithRedis(&cfg.Data.Redis),
		framework.WithServer(&cfg.Data.Server),
	)

	bootstrap.Run(context.Background())
}
```

## API Reference

### Options

- `WithLogger(cfg *logx.Config)` - Configure logger
- `WithDatabase(cfg *database.Config)` - Configure database
- `WithRedis(cfg *redis.Config)` - Configure Redis
- `WithRabbitMQ(cfg *rabbitmq.Config)` - Configure RabbitMQ
- `WithKeycloak(cfg *keycloak.Config)` - Configure Keycloak
- `WithRestClient(opts ...rest.OptionFunc)` - Configure REST client
- `WithScheduler(opts ...scheduler.OptionFunc)` - Configure scheduler
- `WithServer(cfg *server.Config)` - Configure HTTP server

### Lifecycle Methods

- `Init(ctx context.Context) error` - Initialize all services
- `Start(ctx context.Context) error` - Start all services (blocks)
- `Stop(ctx context.Context) error` - Stop all services gracefully
- `Run(ctx context.Context) error` - Init + Start + Stop (convenience method)

### Lifecycle Hooks

- `BeforeInit(fn func(ctx context.Context) error)` - Before initialization
- `AfterInit(fn func(ctx context.Context) error)` - After initialization
- `BeforeStart(fn func(ctx context.Context) error)` - Before starting
- `AfterStart(fn func(ctx context.Context) error)` - After starting
- `BeforeStop(fn func(ctx context.Context) error)` - Before stopping
- `AfterStop(fn func(ctx context.Context) error)` - After stopping

### Getters

- `GetLogger() logx.Logger`
- `GetDatabase() *database.Database`
- `GetRedis() *redis.Cache`
- `GetRabbitMQ() *rabbitmq.RabbitMQ`
- `GetKeycloak() keycloak.KC`
- `GetRest() *rest.Client`
- `GetScheduler() *scheduler.Scheduler`

### Utilities

- `Health(ctx context.Context) map[string]error` - Check health of all services
- `Context() context.Context` - Get bootstrap context
- `Shutdown()` - Trigger graceful shutdown

## Best Practices

1. **Always use context**: Pass context through all operations
2. **Handle errors**: Check errors from Init/Start/Stop
3. **Use hooks**: Use lifecycle hooks for setup/cleanup
4. **Health checks**: Implement health checks for monitoring
5. **Graceful shutdown**: Let bootstrap handle shutdown signals

## Notes

- Services are initialized in the order they are provided
- If a service fails to initialize, it logs an error but continues
- All services are automatically closed during Stop()
- HTTP server runs in a goroutine and blocks until shutdown signal
- Scheduler automatically stops when context is cancelled
