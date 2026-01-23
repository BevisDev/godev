package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockJob struct {
	name   string
	called int32
	panic  bool
	done   chan struct{}
}

func (m *mockJob) Handle(ctx context.Context) {
	atomic.AddInt32(&m.called, 1)

	// báo hiệu job đã chạy
	if m.done != nil {
		select {
		case <-m.done:
		default:
			close(m.done)
		}
	}

	if m.panic {
		panic("boom")
	}
}

func TestScheduler_Start_Idempotent(t *testing.T) {
	s := New()

	s.Register(&Job{
		Name:    "job1",
		Handler: &mockJob{},
		Cron:    "@every 1s",
		IsOn:    true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)
	s.Start(ctx)

	assert.Len(t, s.cron.Entries(), 1)
}

func TestScheduler_Stop_ContextCancel(t *testing.T) {
	s := New()

	job := &mockJob{done: make(chan struct{})}

	s.Register(&Job{
		Name:    "job1",
		Handler: job,
		Cron:    "@every 100ms",
		IsOn:    true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	<-job.done // first run

	cancel()

	called := atomic.LoadInt32(&job.called)
	time.Sleep(300 * time.Millisecond)

	assert.Equal(t, called, atomic.LoadInt32(&job.called))
}

func TestScheduler_RegisterJob_Success(t *testing.T) {
	s := New()
	job := &mockJob{}

	s.Register(&Job{
		Handler: job,
		Cron:    "*/1 * * * *",
		IsOn:    true,
		Name:    "job1",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	entries := s.cron.Entries()
	require.Len(t, entries, 1, "expected 1 cron entry")
}

func TestScheduler_RegisterJob_Disabled(t *testing.T) {
	s := New()

	s.Register(&Job{
		Name:    "job1",
		IsOn:    false,
		Cron:    "*/1 * * * *",
		Handler: &mockJob{name: "job1"},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	assert.Len(t, s.cron.Entries(), 0, "expected no cron entries for disabled job")
}

func TestScheduler_RegisterJob_EmptyCron(t *testing.T) {
	s := New()

	s.Register(&Job{
		Handler: &mockJob{name: "job1"},
		Cron:    "",
		IsOn:    true,
		Name:    "job1",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)

	assert.Len(t, s.cron.Entries(), 0, "expected no cron entries when cron is empty")
}

func TestScheduler_JobPanicRecovered(t *testing.T) {
	s := New()

	job := &mockJob{
		panic: true,
		done:  make(chan struct{}),
	}

	s.Register(&Job{
		Name:    "job1",
		Handler: job,
		Cron:    "@every 1s",
		IsOn:    true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert.NotPanics(t, func() {
		s.Start(ctx)
	})

	select {
	case <-job.done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("job was not executed")
	}

	assert.GreaterOrEqual(t, atomic.LoadInt32(&job.called), int32(1))
}

func TestScheduler_All(t *testing.T) {
	s := New()
	j1 := &mockJob{name: "j1"}
	j2 := &mockJob{name: "j2"}

	s.Register(
		&Job{
			Handler: j1,
			Cron:    "*/1 * * * *",
			IsOn:    true,
			Name:    "job1",
		},
		&Job{
			Handler: j2,
			Cron:    "*/1 * * * *",
			IsOn:    true,
			Name:    "job2",
		},
	)

	all := s.All()
	require.Len(t, all, 2)
	assert.Contains(t, all, "job1")
	assert.Contains(t, all, "job2")
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

	s.Register(
		&Job{
			Handler: j1,
			Cron:    "*/1 * * * *",
			IsOn:    true,
			Name:    "dup",
		},
		&Job{
			Handler: j2,
			Cron:    "*/1 * * * *",
			IsOn:    true,
			Name:    "dup",
		},
	)

	all := s.All()
	require.Len(t, all, 1)
}

func TestScheduler_InvalidCron(t *testing.T) {
	s := New()
	j1 := &mockJob{name: "bad"}

	s.Register(
		&Job{
			Cron:    "invalid cron expression",
			IsOn:    true,
			Handler: j1,
			Name:    "bad",
		},
	)

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
