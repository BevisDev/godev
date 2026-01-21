package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BevisDev/godev/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Run(ctx context.Context, cfg *Config) error {
	if cfg == nil {
		return errors.New("[server] config is nil")
	}
	cfg.withDefaults()

	// Set up signal handling for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	// gin engine
	var r *gin.Engine
	if cfg.IsProduction {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()

		if cfg.Recovery != nil {
			r.Use(gin.CustomRecovery(cfg.Recovery))
		} else {
			r.Use(gin.Recovery())
		}
	} else {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		r = gin.Default()
	}

	if cfg.Setup != nil {
		cfg.Setup(r)
	}

	if len(cfg.Proxies) > 0 {
		_ = r.SetTrustedProxies(cfg.Proxies)
	}

	srv := newHTTPServer(r, cfg)
	errCh := make(chan error, 1)

	// start server
	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("context cancelled")
	case <-sig:
		log.Println("shutdown signal received")
	case err := <-errCh:
		return err
	}

	// graceful shutdown
	shutdownCtx, cancel := utils.NewCtxTimeout(ctx, cfg.ShutdownTimeout)
	defer cancel()

	if cfg.Shutdown != nil {
		if err := cfg.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		return err
	}

	log.Println("server stopped")
	return nil
}

func newHTTPServer(handler http.Handler, cfg *Config) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}
