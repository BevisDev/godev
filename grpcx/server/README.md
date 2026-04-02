# gRPC server (`grpcx/server`)

Package **`server`** (import path `github.com/BevisDev/godev/grpcx/server`) wraps [`google.golang.org/grpc`](https://pkg.go.dev/google.golang.org/grpc) with lifecycle tương tự [`ginfw/server`](../../ginfw/server/README.md): `Start` / `Stop` / `Run`, hook `Setup` / `Shutdown`, graceful stop có timeout.

---

## Defaults

Các giá trị mặc định áp dụng khi field trong `Config` để trống / zero:

- `Network`: `"tcp"`
- `Port`: `9090` (chỉ khi `Host` rỗng và `Port == 0`)
- `ShutdownTimeout`: `15s`

---

## Features

- Graceful shutdown (`GracefulStop`, hết hạn thì `Stop`)
- `Start()` không block; `Run()` chờ signal / `ctx` / lỗi từ `Serve`
- Chuỗi unary + stream interceptor
- `Setup` đăng ký service; `Shutdown` cleanup trước khi tắt server

---

## Config

| Field | Mô tả |
|-------|--------|
| `Network` | Listener network (mặc định: `"tcp"`) |
| `Host` | Listen host (rỗng => listen on all interfaces) |
| `Port` | TCP port để listen |
| `ShutdownTimeout` | Thời gian tối đa cho graceful stop (mặc định: `DefaultShutdownTimeout`) |
| `UnaryInterceptors` | Chain unary interceptor |
| `StreamInterceptors` | Chain stream interceptor |
| `ServerOptions` | Truyền thẳng vào `grpc.NewServer` |
| `Setup` | `func(s *grpc.Server)` — đăng ký service |
| `Shutdown` | `func(ctx context.Context) error` — gọi trước khi shutdown server |

---

## GRPCApp

| API | Mô tả |
|-----|--------|
| `New(cfg *Config) (*GRPCApp, error)` | `cfg == nil` → lỗi |
| `Start() error` | Listen + `Serve` trong goroutine |
| `Stop(ctx context.Context) error` | Hook + graceful stop (theo timeout đã clone trong config) |
| `Run(ctx context.Context) error` | `Start` rồi chờ SIGINT/SIGTERM, `ctx.Done()`, hoặc lỗi |

---

## Quick start

```go
package main

import (
	"context"
	"log"

	grpcsrv "github.com/BevisDev/godev/grpcx/server"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	app, err := grpcsrv.New(&grpcsrv.Config{
		Port: 9090,
		Setup: func(s *grpc.Server) {
			// pb.RegisterYourServiceServer(s, handler)
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

---

## Gợi ý tích hợp `framework`

Module `framework` hiện **chưa** có `WithGRPC` có sẵn. Bạn có thể tự khởi tạo `grpcx/server` trong hook `BeforeStart` / `AddServices`, hoặc chạy song song `Bootstrap.Run` và một goroutine `grpcsrv.New` + `Run` với `context` đời sống app (ví dụ `bootstrap.Context()`).
