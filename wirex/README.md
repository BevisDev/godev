# wirex (Google Wire helpers)

`wirex` is a small helper module for using [`github.com/google/wire`](https://github.com/google/wire) in your
application repo (a "child repo" that embeds/depends on this library).

This module focuses on **reusable provider sets** (`wire.NewSet(...)`) exported from `wirex/sets.go`.
Code generation is done in the child repo, not inside `godev`.

## Why `wirex` exists

- Avoid duplicating provider set definitions across multiple services.
- Keep `godev` as a pure library: app-level dependency graphs live in your repo.

## What this module provides

- `wirex/sets.go`
  - `wirex.DatabaseSet`, `wirex.RedisSet`, `wirex.LoggerSet`, ...
  - `wirex.InfraSet` (grouped infra constructors)
- `wirex/wire.go`
  - a minimal `wireinject` entrypoint so the child repo can import `wirex`.

## Prerequisites (child repo)

1. Add dependency:
   - `go get github.com/google/wire`
2. Install the Wire CLI in your environment (typically via `go install`):
   - `go install github.com/google/wire/cmd/wire@latest`
3. Create a `wireinject` file in your child repo (commonly `app/wire.go`).

## Example: child repo `app/wire.go`

> Put this file inside your application repository (not here).

```go
//go:build wireinject
// +build wireinject

package app

import (
	"context"

	"github.com/BevisDev/godev/database"
	"github.com/BevisDev/godev/grpcx/server" // example, if you use grpcx server
	"github.com/BevisDev/godev/rabbitmq"
	"github.com/BevisDev/godev/wirex"

	"github.com/google/wire"
)

// AppContainer is an example return type for your DI graph.
type AppContainer struct {
	// db *database.DB
	// rabbit *rabbitmq.MQ
	// grpc *server.GRPCApp
}

func InitApp(
	ctx context.Context,
	dbCfg *database.Config,
	rabbitCfg *rabbitmq.Config,
	grpcCfg *server.Config,
) (*AppContainer, error) {
	wire.Build(
		// Provider sets from this repo:
		wirex.DatabaseSet,

		// Example: you often need "thin wrappers" for constructors that take
		// context.Context or variadic options.
		ProvideRabbit,
		server.New, // grpcx/server.New has signature (*Config) (*GRPCApp, error)

		// Constructors/wrappers that wire.Build needs to produce *AppContainer
		ProvideAppContainer,
	)

	return nil, nil
}

// ProvideRabbit is a thin wrapper around rabbitmq.New so it matches Wire's needs.
func ProvideRabbit(ctx context.Context, cfg *rabbitmq.Config) (*rabbitmq.MQ, error) {
	return rabbitmq.New(ctx, cfg)
}

func ProvideAppContainer(/* inject what you need */) *AppContainer {
	return &AppContainer{}
}
```

Then run Wire in the child repo:

```sh
wire ./path/to/your/wireinject
```

## Notes / common gotchas

- `wirex` provider sets in `wirex/sets.go` assume your injector supplies the corresponding `*X.Config` values.
- Constructors that take `context.Context`, `...Option` / variadic options, or framework-specific `framework.Option`
  usually need a **thin wrapper** defined in the child repo.
- This module does **not** commit generated `wire_gen.go` files.
  Keep generated code in the child repo.

## Related

- Wire docs: https://github.com/google/wire
