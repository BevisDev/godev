package logger

// Config defines the configuration options for setting up the application logger.
//
// It supports file-based logging with rotation (via lumberjack) and optional
// profile-based behavior (e.g., dev/prod).
type Config struct {
	// IsProduction indicates whether the application is running in PROD environment.
	IsProduction bool

	// IsLocal indicates whether the application is running in DEV environment.
	IsLocal bool

	// MaxSize is the maximum size (in megabytes) of the log file before it gets rotated.
	MaxSize int

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int

	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int

	// Compress determines whether rotated log files are compressed using gzip.
	Compress bool

	// IsRotate indicates whether to rotate log files by day or by module (depending on implementation).
	IsRotate bool

	// Cron defines the time-based rotation schedule (cron format).
	// Example: "0 0 * * *" rotates logs daily at midnight.
	Cron string

	// DirName is the directory path where logs will be stored.
	DirName string

	// Filename is the base name of the log file (e.g., "app.log").
	Filename string

	// CallerConfig controls zap caller skip levels for request/response logging.
	CallerConfig CallerConfig
}

type CallerConfig struct {
	Request  SkipGroup // Request defines caller skip config for internal/external request logs.
	Response SkipGroup // Response defines caller skip config for internal/external response logs.
}

// SkipGroup holds the caller skip values for internal and external contexts.
type SkipGroup struct {
	Internal int // Internal number of caller frames to skip for internal log calls.
	External int // External number of caller frames to skip for external log calls.
}

func (c *Config) clone() *Config {
	clone := *c
	if clone.MaxSize <= 0 {
		clone.MaxSize = 100
	}
	if clone.MaxBackups <= 0 {
		clone.MaxBackups = 100
	}
	if clone.MaxAge <= 0 {
		clone.MaxAge = 30
	}
	if !clone.IsLocal {
		clone.Compress = true
	}
	if clone.Cron == "" {
		clone.Cron = "0 0 * * *"
	}
	if clone.DirName == "" {
		clone.DirName = "./logs"
	}
	if clone.Filename == "" {
		clone.Filename = "app.log"
	}
	if clone.CallerConfig.Request.Internal <= 0 {
		clone.CallerConfig.Request.Internal = 2
	}
	if clone.CallerConfig.Request.External <= 0 {
		clone.CallerConfig.Request.External = 5
	}
	if clone.CallerConfig.Response.Internal <= 0 {
		clone.CallerConfig.Response.Internal = 2
	}
	if clone.CallerConfig.Response.External <= 0 {
		clone.CallerConfig.Response.External = 6
	}
	return &clone
}
