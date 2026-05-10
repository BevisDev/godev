package providers

import (
	"context"

	"github.com/BevisDev/godev/scheduler"
)

type SchedulerProvider struct {
	opts []scheduler.Option
	s    *scheduler.Scheduler
}

func NewSchedulerProvider(opts ...scheduler.Option) *SchedulerProvider {
	return &SchedulerProvider{opts: opts}
}

func (p *SchedulerProvider) Init(ctx context.Context) error {
	_ = ctx
	p.s = scheduler.New(p.opts...)
	return nil
}

func (p *SchedulerProvider) Start(ctx context.Context) error {
	if p.s != nil {
		p.s.Start(ctx)
	}
	return nil
}

func (p *SchedulerProvider) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *SchedulerProvider) Scheduler() *scheduler.Scheduler {
	return p.s
}
