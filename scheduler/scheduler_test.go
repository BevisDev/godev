package scheduler

import (
	"context"
	"testing"
	"time"
)

type mockJob struct {
	name   string
	called int
	panic  bool
}

func (m *mockJob) Name() string {
	return m.name
}

func (m *mockJob) Handle(ctx context.Context) {
	m.called++
	if m.panic {
		panic("boom")
	}
}

func TestScheduler_RegisterJob_Success(t *testing.T) {
	s := NewScheduler()

	job := &mockJob{name: "job1"}

	s.Register(job, JobConfig{
		Cron: "*/1 * * * *",
		IsOn: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)

	entries := s.cron.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 cron entry, got %d", len(entries))
	}
}

func TestScheduler_RegisterJob_Disabled(t *testing.T) {
	s := NewScheduler()

	s.Register(&mockJob{name: "job1"}, JobConfig{
		Cron: "*/1 * * * *",
		IsOn: false,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)

	if len(s.cron.Entries()) != 0 {
		t.Fatal("expected no cron entries for disabled job")
	}
}

func TestScheduler_RegisterJob_EmptyCron(t *testing.T) {
	s := NewScheduler()

	s.Register(&mockJob{name: "job1"}, JobConfig{
		Cron: "",
		IsOn: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)

	if len(s.cron.Entries()) != 0 {
		t.Fatal("expected no cron entries when cron is empty")
	}
}

func TestScheduler_JobPanicRecovered(t *testing.T) {
	s := NewScheduler(WithSeconds())

	job := &mockJob{
		name:  "panic-job",
		panic: true,
	}

	s.Register(job, JobConfig{
		Cron: "*/1 * * * * *",
		IsOn: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// should not panic
	s.Start(ctx)

	time.Sleep(1100 * time.Millisecond)

	if job.called == 0 {
		t.Fatal("expected job to be called at least once")
	}
}
