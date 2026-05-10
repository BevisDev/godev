package framework

import (
	"context"
)

type Initializer struct {
	phase *Phase
	b     *Bootstrap
}

func NewInitializer(b *Bootstrap) *Initializer {
	return &Initializer{
		phase: NewPhase(StateInit),
		b:     b,
	}
}

// Run initializes all configured services.
func (i *Initializer) Run(ctx context.Context) error {
	if i.phase.isDone() {
		return ErrAlreadyInitialized
	}

	if err := i.phase.runPre(ctx); err != nil {
		return err
	}

	if i.b.svc == nil {
		i.b.svc = NewService(ctx, i.b.svcConf)
	}

	if err := i.b.svc.Run(ctx); err != nil {
		return err
	}

	if err := i.phase.runPost(ctx); err != nil {
		return err
	}

	i.phase.markDone()

	return nil
}

func (i *Initializer) Before(fn Hook) error {
	return i.phase.Before(fn)
}

func (i *Initializer) After(fn Hook) error {
	return i.phase.After(fn)
}
