package framework

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BevisDev/godev/ginfw/server"
)

type Starter struct {
	phase *Phase
	b     *Bootstrap
}

func NewStarter(b *Bootstrap) *Starter {
	return &Starter{
		phase: NewPhase(StateStart),
		b:     b,
	}
}

// Before registers a hook to run before starting services.
func (s *Starter) Before(fn Hook) error {
	return s.phase.Before(fn)
}

// After registers a hook to run after starting services.
func (s *Starter) After(fn Hook) error {
	return s.phase.After(fn)
}

// Start starts all services and blocks until shutdown signal.
func (s *Starter) Start(ctx context.Context) error {
	if s.phase.isDone() {
		return ErrAlreadyStarted
	}

	if err := s.phase.runPre(ctx); err != nil {
		return err
	}

	s.b.log.Info("starting services...")

	// Start scheduler if configured
	if svc := s.b.svc; svc != nil && svc.scheduler != nil {
		svc.scheduler.Start(ctx)
	}

	if svc := s.b.svc; svc != nil && svc.rabbitmq != nil && svc.rabbitmq.Consumer() != nil {
		go svc.rabbitmq.Consumer().Start(ctx)
	}

	// Start Kafka consumer if configured (handler registered and consumer initialized)
	if svc := s.b.svc; svc != nil && svc.kafka != nil && svc.kafka.HasConsumer() && svc.cfg.kafkaConsumerHandler != nil {
		handler := svc.cfg.kafkaConsumerHandler
		if svc.cfg.kafkaConsumerRetry.enabled {
			maxRetries := svc.cfg.kafkaConsumerRetry.maxRetries
			retryDelay := svc.cfg.kafkaConsumerRetry.retryDelay
			go func() {
				_ = svc.kafka.ConsumeWithRetry(ctx, handler, maxRetries, retryDelay)
			}()
		} else {
			go func() {
				_ = svc.kafka.Consume(ctx, handler)
			}()
		}
		s.b.log.Info("Kafka consumer started")
	}

	// Start HTTP server if configured
	if s.b.serverConf != nil {
		s.b.httpApp = server.New(s.b.serverConf)
		if err := s.b.httpApp.Start(); err != nil {
			return fmt.Errorf("[bootstrap] failed to start HTTP server: %w", err)
		}
	}

	if err := s.phase.runPost(ctx); err != nil {
		return err
	}
	s.phase.markDone()

	s.b.mu.Lock()
	s.b.started = true
	s.b.mu.Unlock()

	// Setup signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		s.b.log.Info("root context cancelled")
	case sig := <-sig:
		s.b.log.Info("received signal: %v", sig)
	}

	return nil
}
