package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BevisDev/godev/consts"
	"github.com/BevisDev/godev/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type HTTPApp struct {
	config *Config
	engine *gin.Engine
	server *http.Server
	errCh  chan error
}

// New creates a new HTTPApp instance with the provided configuration.
// It initializes the Gin engine, applies configuration, and sets up the HTTP server.
func New(cfg *Config) *HTTPApp {
	config := cfg.clone()

	// Initialize Gin engine based on production mode
	var r *gin.Engine
	if config.IsProduction {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()

		if config.Recovery != nil {
			r.Use(gin.CustomRecovery(config.Recovery))
		} else {
			r.Use(gin.Recovery())
		}
	} else {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		r = gin.Default()
	}

	// Apply setup hook if provided
	if config.Setup != nil {
		config.Setup(r)
	}

	// Configure trusted proxies
	if len(config.Proxies) > 0 {
		_ = r.SetTrustedProxies(config.Proxies)
	}

	srv := newHTTPServer(r, config)

	return &HTTPApp{
		config: config,
		engine: r,
		server: srv,
		errCh:  make(chan error, 1),
	}
}

// Start starts the HTTP server in a goroutine.
// Returns immediately after starting the server.
// Use Run() to start and wait for shutdown signals.
func (h *HTTPApp) Start() error {
	go func() {
		log.Printf("[server] listening on :%s", h.config.Port)
		if err := h.server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			h.errCh <- err
		}
	}()
	return nil
}

func newHTTPServer(handler http.Handler, config *Config) *http.Server {
	return &http.Server{
		Addr:              consts.Colon + config.Port,
		Handler:           handler,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
	}
}

// Stop gracefully stops the HTTP server.
// It calls the Shutdown hook (if configured) and then shuts down the HTTP server.
// The context is used to control the shutdown timeout.
func (h *HTTPApp) Stop(ctx context.Context) error {
	shutdownCtx, cancel := utils.NewCtxTimeout(ctx, h.config.ShutdownTimeout)
	defer cancel()

	// Call custom shutdown hook if provided
	if h.config.Shutdown != nil {
		if err := h.config.Shutdown(shutdownCtx); err != nil {
			log.Printf("[server] shutdown hook error: %v", err)
		}
	}

	// Shutdown HTTP server
	if err := h.server.Shutdown(shutdownCtx); err != nil {
		_ = h.server.Close()
		return err
	}

	log.Println("[server] stopped")
	return nil
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
// It handles SIGINT and SIGTERM signals for graceful shutdown.
// Returns an error if the server fails to start or encounters an error.
func (h *HTTPApp) Run(ctx context.Context) error {
	// Set up signal handling for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	// Start server
	if err := h.Start(); err != nil {
		return err
	}

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		log.Println("[server] root context cancelled")
	case s := <-sig:
		log.Printf("[server] received signal %v", s)
	case err := <-h.errCh:
		return err
	}

	// Graceful shutdown
	return h.Stop(ctx)
}
