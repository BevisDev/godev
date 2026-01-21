## ⚙️ Scheduler Integration

This library wraps `robfig/cron` with a higher-level `Scheduler` abstraction
to simplify cron job management in Go applications.

The scheduler provides:

- Job registration with explicit configuration
- Per-job enable / disable control
- Panic-safe job execution
- Timezone-aware scheduling
- Graceful shutdown via `context.Context`

---

## ✨ Features

- Register jobs using cron expressions
- Enable or disable jobs at runtime via configuration
- Optional support for cron expressions with seconds
- Explicit timezone configuration
- Panic recovery per job (scheduler never crashes)
- Graceful shutdown using `context.Context`
- No global mutable state (safe for multi-project usage)

---

## 📦 Example

```go
import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/BevisDev/godev/scheduler"
)

type HelloJob struct{}

func NewHelloJob() scheduler.Job {
    return &HelloJob{}
}

func (j *HelloJob) Name() string {
    return "hello-job"
}

func (j *HelloJob) Handle(ctx context.Context) {
    log.Println("[job] hello world")
}

func main() {
    // Create root context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle OS signals for graceful shutdown
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

    // Initialize scheduler
    s := scheduler.New(
        scheduler.WithSeconds(),
        scheduler.WithTimezone("Asia/Ho_Chi_Minh"),
    )

    // Register jobs
    s.Register(
        NewHelloJob(),
        scheduler.JobConfig{
            Cron: "*/5 * * * * *", // every 5 seconds
            IsOn: true,
        },
    )

    // Start scheduler
    s.Start(ctx)

    log.Println("[main] scheduler started")

    // Wait for shutdown signal
    <-sig
    log.Println("[main] shutting down...")
    
    cancel()
    time.Sleep(time.Second) // allow cron to stop gracefully
}

```

**Cron Expression Format:**

| Field        | Mandatory | Allowed Values  | Special Characters |
|--------------|-----------|-----------------|--------------------|
| Seconds      | Yes       | 0-59            | `* / , -`          |
| Minutes      | Yes       | 0-59            | `* / , -`          |
| Hours        | Yes       | 0-23            | `* / , -`          |
| Day of Month | Yes       | 1-31            | `* / , - ?`        |
| Month        | Yes       | 1-12 or JAN-DEC | `* / , -`          |
| Day of Week  | Yes       | 0-6 or SUN-SAT  | `* / , - ?`        |
