package interfaces

import "context"

type Loader interface {
	Load(ctx context.Context) error
}

type Starter interface {
	Start(ctx context.Context) error
}

type Stopper interface {
	Stop(ctx context.Context) error
}

type Closer interface {
	Close(ctx context.Context) error
}

type Initializer interface {
	Init(ctx context.Context) error
}

type Register interface {
	Register(ctx context.Context)
}

type Runner interface {
	Run(ctx context.Context) error
}

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
