package framework

import (
	"context"
)

type Stopper struct {
	phase *Phase
	b     *Bootstrap
}

func NewStopper(b *Bootstrap) *Stopper {
	return &Stopper{
		phase: NewPhase(StateStop),
		b:     b,
	}
}

// Before registers a hook to run before stopping services.
func (s *Stopper) Before(fn Hook) error {
	return s.phase.Before(fn)
}

// After registers a hook to run after stopping services.
func (s *Stopper) After(fn Hook) error {
	return s.phase.After(fn)
}

// Stop gracefully stops all services.
func (b *Bootstrap) Stop(ctx context.Context) error {
	b.mu.Lock()
	if !b.started {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	// Cancel bootstrap context so Kafka consumer and other goroutines using b.ctx exit
	b.cancel()

	b.log.Info("stopping services...")

	if err := b.stopper.phase.runPre(ctx); err != nil {
		return err
	}

	// Stop HTTP server if configured
	if b.httpApp != nil {
		if err := b.httpApp.Stop(ctx); err != nil {
			b.log.Info("HTTP server stop error: %v", err)
		}
	}

	// Close services
	b.closeServices()

	if err := b.stopper.phase.runPost(ctx); err != nil {
		return err
	}
	b.stopper.phase.markDone()

	b.mu.Lock()
	b.started = false
	b.mu.Unlock()

	b.log.Info("all services stopped")
	return nil
}
