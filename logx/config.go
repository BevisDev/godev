package logx

// Config defines the configuration options for setting up the application logger.
//
// It supports file-based logging with rotation (via lumberjack) and optional
// profile-based behavior (e.g., dev/prod).
type Config struct {
	// Profile indicates the runtime profile (e.g., "dev", "prod") and can affect logging format/output
	Profile string

	// MaxSize is the maximum size (in megabytes) of the log file before it gets rotated.
	MaxSize int

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int

	// MaxAge is the maximum number of days to retain old log files.
	MaxAge int

	// Compress determines whether rotated log files are compressed using gzip.
	Compress bool

	// IsSplit indicates whether to split log files by day or by module (depending on implementation).
	IsSplit bool

	// DirName is the directory path where logs will be stored.
	DirName string

	// Filename is the base name of the log file (e.g., "app.log").
	Filename string

	// CallerConfig defines the number of caller stack frames to skip
	// when logging for different request/response contexts.
	// Useful for configuring zap.AddCallerSkip(...) dynamically.
	CallerConfig CallerConfig
}

type CallerConfig struct {
	// Request defines caller skip config for internal/external request logs.
	Request SkipGroup

	// Response defines caller skip config for internal/external response logs.
	Response SkipGroup
}

// SkipGroup holds the caller skip values for internal and external contexts.
type SkipGroup struct {
	// Internal number of caller frames to skip for internal log calls.
	Internal int

	// External number of caller frames to skip for external log calls.
	External int
}
