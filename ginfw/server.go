package ginfw

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(ctx context.Context, opt Options) error {
	opt.withDefault()

	// Set up signal handling for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	// gin engine
	var r *gin.Engine
	if opt.IsProduction {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(gin.Recovery())
	} else {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		r = gin.Default()
	}

	if opt.Setup != nil {
		opt.Setup(r)
	}

	if len(opt.Proxies) > 0 {
		_ = r.SetTrustedProxies(opt.Proxies)
	}

	srv := newServer(":"+opt.Port, r, opt)
	errCh := make(chan error, 1)

	// start server
	go func() {
		log.Printf("[ginfw] server listening on :%s", opt.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("[ginfw] context cancelled")
	case <-stop:
		log.Println("[ginfw] shutdown signal received")
	case err := <-errCh:
		return err
	}

	// graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, opt.ShutdownTimeout)
	defer cancel()

	if opt.Shutdown != nil {
		if err := opt.Shutdown(shutdownCtx); err != nil {
			log.Printf("[ginfw] func shutdown error: %v", err)
		}
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		return err
	}

	log.Println("[ginfw] server stopped")
	return nil
}

func newServer(addr string, handler http.Handler, opt Options) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
		IdleTimeout:  opt.IdleTimeout,
	}
}
