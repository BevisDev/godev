package scheduler

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/console"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	*options
	cron    *cron.Cron
	jobs    map[string]*Job
	started bool
	mu      sync.Mutex
	log     *console.Logger
}

func New(opts ...Option) *Scheduler {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	cronOpts := []cron.Option{
		cron.WithLocation(options.location),
	}
	if options.useSeconds {
		cronOpts = append(cronOpts, cron.WithSeconds())
	}

	return &Scheduler{
		options: options,
		cron:    cron.New(cronOpts...),
		jobs:    make(map[string]*Job),
		log:     console.New("scheduler"),
	}
}

func (s *Scheduler) Register(jobs ...*Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range jobs {
		if job == nil || job.Name == "" || job.Cron == "" || job.Handler == nil {
			continue
		}

		if _, ok := s.jobs[job.Name]; ok {
			s.log.Info("job %s already registered, override", job.Name)
		}

		s.jobs[job.Name] = job
	}
}

func (s *Scheduler) All() map[string]*Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make(map[string]*Job, len(s.jobs))
	for k, v := range s.jobs {
		cp[k] = v
	}

	return cp
}

func (s *Scheduler) Timezone() string {
	return s.cron.Location().String()
}

// register iterates over all registered jobs and schedules enabled ones
// based on their cron configuration.
// It safely wraps job execution with panic recovery.
func (s *Scheduler) register() {
	s.mu.Lock()
	jobs := make(map[string]*Job, len(s.jobs))
	for k, v := range s.jobs {
		jobs[k] = v
	}
	s.mu.Unlock()

	for k, v := range jobs {
		name := k
		job := v

		if !job.IsOn {
			s.log.Info("job %s is disabled", name)
			continue
		}

		_, err := s.cron.AddFunc(job.Cron, func() {
			ctx := utils.NewCtx()

			defer func() {
				if r := recover(); r != nil {
					s.log.Error("[RECOVER] job %s: %v \npanic: %s",
						name, r, debug.Stack(),
					)
				}
			}()

			job.Handler.Handle(ctx)
		})
		if err != nil {
			s.log.Error("error register job %s: %v", name, err)
		}
	}
}

// Start registers all jobs, starts the cron scheduler,
// and stops it gracefully when the context is canceled.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}

	s.started = true
	s.mu.Unlock()

	s.register()

	if len(s.cron.Entries()) == 0 {
		s.log.Info("no jobs registered")
		return
	}

	s.cron.Start()
	s.log.Info("started successfully, timezone=%s",
		s.Timezone(),
	)

	go func() {
		<-ctx.Done()
		s.log.Info("stopping...")
		s.cron.Stop()
	}()
}
