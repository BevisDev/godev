package server

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// Config defines the configuration for running a gRPC server.
type Config struct {
	// Network is the listener network. Typical value: "tcp".
	Network string

	// Address is the listen address (host:port).
	// Examples: ":9090", "127.0.0.1:9090".
	Address string

	// ShutdownTimeout is the maximum duration for graceful shutdown.
	ShutdownTimeout time.Duration

	// UnaryInterceptors are chained and applied in the provided order.
	UnaryInterceptors []grpc.UnaryServerInterceptor

	// StreamInterceptors are chained and applied in the provided order.
	StreamInterceptors []grpc.StreamServerInterceptor

	// ServerOptions are passed directly to grpc.NewServer.
	ServerOptions []grpc.ServerOption

	// Setup is an optional hook to register services on grpc.Server.
	Setup func(s *grpc.Server)

	// Shutdown is an optional hook invoked during graceful shutdown.
	Shutdown func(ctx context.Context) error
}

func (c *Config) clone() *Config {
	cc := *c
	if cc.Network == "" {
		cc.Network = "tcp"
	}
	if cc.Address == "" {
		cc.Address = ":9090"
	}
	if cc.ShutdownTimeout <= 0 {
		cc.ShutdownTimeout = 15 * time.Second
	}

	if len(cc.UnaryInterceptors) > 0 {
		cc.UnaryInterceptors = append([]grpc.UnaryServerInterceptor(nil), cc.UnaryInterceptors...)
	}
	if len(cc.StreamInterceptors) > 0 {
		cc.StreamInterceptors = append([]grpc.StreamServerInterceptor(nil), cc.StreamInterceptors...)
	}
	if len(cc.ServerOptions) > 0 {
		cc.ServerOptions = append([]grpc.ServerOption(nil), cc.ServerOptions...)
	}

	return &cc
}
