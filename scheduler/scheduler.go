package scheduler

import (
	"context"
	"log"
	"runtime/debug"

	"github.com/BevisDev/godev/utils"
	"github.com/robfig/cron/v3"
)

type JobEntry struct {
	Job    Job
	Config JobConfig
}

type Scheduler struct {
	cron *cron.Cron
	jobs map[string]*JobEntry
}

func New(fs ...OptionFunc) *Scheduler {
	options := defaultOptions()
	for _, f := range fs {
		f(options)
	}

	cronOpts := []cron.Option{
		cron.WithLocation(options.Location),
	}
	if options.WithSeconds {
		cronOpts = append(cronOpts, cron.WithSeconds())
	}

	return &Scheduler{
		cron: cron.New(cronOpts...),
		jobs: make(map[string]*JobEntry),
	}
}

func (s *Scheduler) Register(jobs ...*JobEntry) {
	for _, item := range jobs {
		name := item.Job.Name()
		if _, ok := s.jobs[name]; ok {
			log.Printf("[scheduler] job %s already registered, override", name)
		}
		s.jobs[name] = item
	}
}

func (s *Scheduler) All() map[string]*JobEntry {
	return s.jobs
}

func (s *Scheduler) Timezone() string {
	return s.cron.Location().String()
}

// register iterates over all registered jobs and schedules enabled ones
// based on their cron configuration.
// It safely wraps job execution with panic recovery.
func (s *Scheduler) register() {
	for name, entry := range s.All() {
		cfg := entry.Config
		jobName := name
		job := entry.Job

		if cfg.Cron == "" {
			log.Printf("[scheduler] cron %s not found in config", jobName)
			continue
		}

		if !cfg.IsOn {
			log.Printf("[scheduler] cron %s is disabled", jobName)
			continue
		}

		_, err := s.cron.AddFunc(cfg.Cron, func() {
			ctx := utils.NewCtx()

			defer func() {
				if r := recover(); r != nil {
					log.Printf(
						"[RECOVER] job %s: %v\n%s",
						jobName, r, debug.Stack(),
					)
				}
			}()

			job.Handle(ctx)
		})
		if err != nil {
			log.Printf("[scheduler] error register cron %s: %v", jobName, err)
		}
	}
}

// Start registers all jobs, starts the cron scheduler,
// and stops it gracefully when the context is canceled.
func (s *Scheduler) Start(ctx context.Context) {
	s.register()

	if len(s.cron.Entries()) == 0 {
		log.Println("[scheduler] no jobs registered")
		return
	}

	s.cron.Start()
	log.Printf(
		"[scheduler] started successfully, timezone=%s",
		s.Timezone(),
	)

	go func() {
		<-ctx.Done()
		log.Println("[scheduler] stopping...")
		s.cron.Stop()
	}()
}
