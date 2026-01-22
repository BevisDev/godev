package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	s := New()
	job := &mockJob{name: "job1"}

	s.Register(job, JobConfig{
		Cron: "*/1 * * * *",
		IsOn: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	entries := s.cron.Entries()
	require.Len(t, entries, 1, "expected 1 cron entry")
}

func TestScheduler_RegisterJob_Disabled(t *testing.T) {
	s := New()

	s.Register(&mockJob{name: "job1"}, JobConfig{
		Cron: "*/1 * * * *",
		IsOn: false,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	assert.Len(t, s.cron.Entries(), 0, "expected no cron entries for disabled job")
}

func TestScheduler_RegisterJob_EmptyCron(t *testing.T) {
	s := New()

	s.Register(&mockJob{name: "job1"}, JobConfig{
		Cron: "",
		IsOn: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	assert.Len(t, s.cron.Entries(), 0, "expected no cron entries when cron is empty")
}

func TestScheduler_JobPanicRecovered(t *testing.T) {
	s := New(WithSeconds())
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
	assert.NotPanics(t, func() { s.Start(ctx) })

	time.Sleep(1100 * time.Millisecond)
	assert.GreaterOrEqual(t, job.called, 1, "expected job to be called at least once")
}

func TestScheduler_All(t *testing.T) {
	s := New()
	j1 := &mockJob{name: "j1"}
	j2 := &mockJob{name: "j2"}

	s.Register(j1, JobConfig{Cron: "0 * * * *", IsOn: true})
	s.Register(j2, JobConfig{Cron: "0 * * * *", IsOn: true})

	all := s.All()
	require.Len(t, all, 2)
	assert.Contains(t, all, "j1")
	assert.Contains(t, all, "j2")
	assert.Same(t, j1, all["j1"].Job)
	assert.Same(t, j2, all["j2"].Job)
}

func TestScheduler_Timezone(t *testing.T) {
	s := New()
	assert.Contains(t, s.Timezone(), "UTC")

	s2 := New(WithTimezone("UTC"))
	assert.Contains(t, s2.Timezone(), "UTC")
}

func TestScheduler_Register_DuplicateOverride(t *testing.T) {
	s := New()
	j1 := &mockJob{name: "dup"}
	j2 := &mockJob{name: "dup"}

	s.Register(j1, JobConfig{Cron: "0 * * * *", IsOn: true})
	s.Register(j2, JobConfig{Cron: "0 * * * *", IsOn: true})

	all := s.All()
	require.Len(t, all, 1)
	assert.Same(t, j2, all["dup"].Job, "second register overrides first")
}

func TestScheduler_InvalidCron(t *testing.T) {
	s := New()
	s.Register(&mockJob{name: "bad"}, JobConfig{
		Cron: "invalid cron expression",
		IsOn: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	assert.Len(t, s.cron.Entries(), 0, "invalid cron should not add any entry")
}

func TestScheduler_Start_NoJobsRegistered(t *testing.T) {
	s := New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	assert.NotPanics(t, func() { s.Start(ctx) })
	assert.Len(t, s.cron.Entries(), 0)
}
