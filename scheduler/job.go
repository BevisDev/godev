package scheduler

import "context"

type Job interface {
	Name() string
	Handle(ctx context.Context)
}

type JobConfig struct {
	Cron string // cron expression
	IsOn bool   // enable / disable job
}
