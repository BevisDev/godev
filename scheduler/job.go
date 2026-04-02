package scheduler

import "context"

type Handler interface {
	Handle(ctx context.Context)

	// JobName returns the unique name used to register this job in Scheduler.
	JobName() string
}

type Job struct {
	Handler Handler
	Cron    string // cron expression
	IsOn    bool   // enable / disable job
}
