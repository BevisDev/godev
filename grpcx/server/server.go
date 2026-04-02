package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"google.golang.org/grpc"
)

// GRPCApp manages gRPC server lifecycle.
type GRPCApp struct {
	cf     *Config
	server *grpc.Server
	errCh  chan error

	mu  sync.Mutex
	lis net.Listener
}

// New creates a new GRPCApp instance with the provided configuration.
// Returns an error if cf is nil.
func New(cf *Config) (*GRPCApp, error) {
	if cf == nil {
		return nil, fmt.Errorf("[grpcx/server] config is nil")
	}
	cfg := cf.clone()

	opts := append([]grpc.ServerOption(nil), cfg.ServerOptions...)
	if len(cfg.UnaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(cfg.UnaryInterceptors...))
	}
	if len(cfg.StreamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(cfg.StreamInterceptors...))
	}

	s := grpc.NewServer(opts...)
	if cfg.Setup != nil {
		cfg.Setup(s)
	}

	return &GRPCApp{
		cf:     cfg,
		server: s,
		errCh:  make(chan error, 1),
	}, nil
}

// Start starts the gRPC server in a goroutine.
// Returns immediately after the server starts listening.
func (g *GRPCApp) Start() error {
	lis, err := net.Listen(g.cf.Network, consts.Colon+g.cf.Address)
	if err != nil {
		return err
	}

	g.mu.Lock()
	g.lis = lis
	g.mu.Unlock()

	go func() {
		log.Printf("[grpc] listening on %s/%s", g.cf.Network, lis.Addr().String())
		if err := g.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			select {
			case g.errCh <- err:
			default:
			}
		}
	}()

	return nil
}

// Stop gracefully stops the gRPC server.
func (g *GRPCApp) Stop(ctx context.Context) error {
	shutdownCtx, cancel := utils.NewCtxTimeout(ctx, g.cf.ShutdownTimeout)
	defer cancel()

	// Call custom shutdown hook if provided.
	if g.cf.Shutdown != nil {
		if err := g.cf.Shutdown(shutdownCtx); err != nil {
			log.Printf("[grpc] shutdown hook error: %v", err)
		}
	}

	done := make(chan struct{})
	go func() {
		g.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		// Force stop if graceful shutdown exceeds timeout.
		g.server.Stop()
		<-done
	}

	g.mu.Lock()
	lis := g.lis
	g.lis = nil
	g.mu.Unlock()
	if lis != nil {
		_ = lis.Close()
	}

	log.Println("[grpc] stopped")
	return nil
}

// Run starts the gRPC server and blocks until shutdown signal, context cancel, or server error.
func (g *GRPCApp) Run(ctx context.Context) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	if err := g.Start(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		log.Println("[grpc] root context cancelled")
	case s := <-sig:
		log.Printf("[grpc] received signal %v", s)
	case err := <-g.errCh:
		return err
	}

	return g.Stop(ctx)
}
