package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

// Defaults applied by Config.clone when fields are zero.
const (
	defaultListenNetwork   = "tcp"
	defaultListenPort      = 9090
	defaultShutdownTimeout = 15 * time.Second
)

// Config defines the configuration for running a gRPC server.
type Config struct {
	// Network is the listener network. Typical value: "tcp".
	Network string

	// Host is the listen host.
	// - empty => listen on all interfaces (":port")
	// - non-empty => listen on "host:port"
	Host string

	// Port is the TCP port the gRPC server listens on.
	// When Host is empty and Port is 0, it defaults to defaultListenPort.
	// You can set Port=0 with Host non-empty to let the OS pick an ephemeral port.
	Port int

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

func (c *Config) listenAddr() string {
	if c.Host == "" {
		return fmt.Sprintf(":%d", c.Port)
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) clone() *Config {
	cc := *c
	if cc.Network == "" {
		cc.Network = defaultListenNetwork
	}

	if cc.Port == 0 {
		cc.Port = defaultListenPort
	}
	if cc.ShutdownTimeout <= 0 {
		cc.ShutdownTimeout = defaultShutdownTimeout
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
