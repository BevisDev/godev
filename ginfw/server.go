package ginfw

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Run(ctx context.Context, opt *Options) {
	// Set up signal handling for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var r *gin.Engine
	if strings.HasPrefix(opt.Profile, "prod") {
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

	srv := &http.Server{
		Addr:         ":" + opt.Port,
		Handler:      r,
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
		IdleTimeout:  opt.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down…")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if opt.Shutdown != nil {
		opt.Shutdown()
	}

	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()
	}

	log.Println("server stopped")
}
