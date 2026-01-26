package scheduler

import "context"

type Handler interface {
	Handle(ctx context.Context)
}

type Job struct {
	Name    string // name job
	Cron    string // cron expression
	IsOn    bool   // enable / disable job
	Handler Handler
}
