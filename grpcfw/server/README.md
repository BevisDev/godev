# gRPC Server Package (`grpcfw/server`)

The `grpcfw/server` package provides a production-ready gRPC server with graceful shutdown, lifecycle management, and framework-friendly setup hooks.

---

## Features

- ✅ **gRPC Native**: Built on top of `google.golang.org/grpc`
- ✅ **Graceful Shutdown**: Supports timeout-based graceful stop with fallback force stop
- ✅ **Lifecycle Management**: Separate `Start()` and `Stop()` methods
- ✅ **Interceptors Support**: Unary + stream interceptor chaining
- ✅ **Signal Handling**: `Run()` listens for SIGINT/SIGTERM
- ✅ **Framework Integration**: Easy integration with `framework.WithGRPCServer`

---

## Structure

### `Config`

Configuration for the gRPC server:

| Field | Type | Description |
|---|---|---|
| `Network` | `string` | Listener network (default: `"tcp"`) |
| `Address` | `string` | Listen address (default: `":9090"`) |
| `ShutdownTimeout` | `time.Duration` | Graceful shutdown timeout (default: `15s`) |
| `UnaryInterceptors` | `[]grpc.UnaryServerInterceptor` | Unary interceptors chain |
| `StreamInterceptors` | `[]grpc.StreamServerInterceptor` | Stream interceptors chain |
| `ServerOptions` | `[]grpc.ServerOption` | Additional server options |
| `Setup` | `func(s *grpc.Server)` | Hook to register services |
| `Shutdown` | `func(ctx context.Context) error` | Hook for custom cleanup |

### `GRPCApp`

Main server struct:

| Method | Description |
|---|---|
| `New(cfg *Config) *GRPCApp` | Create server instance |
| `Start() error` | Start server in goroutine |
| `Stop(ctx context.Context) error` | Gracefully stop server |
| `Run(ctx context.Context) error` | Start and wait for signal/context |

---

## Quick Start

```go
package main

import (
	"context"
	"log"

	grpcserver "github.com/BevisDev/godev/grpcfw/server"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	app := grpcserver.New(&grpcserver.Config{
		Address: ":9090",
		Setup: func(s *grpc.Server) {
			// Register your generated gRPC services here.
			// pb.RegisterGreeterServer(s, greeterHandler)
		},
	})

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

---

## Integration with Framework

```go
bootstrap := framework.New(
	framework.WithGRPCServer(&grpcserver.Config{
		Address: ":9090",
		Setup: func(s *grpc.Server) {
			// pb.RegisterGreeterServer(s, handler)
		},
	}),
)

bootstrap.Run(context.Background())
```

