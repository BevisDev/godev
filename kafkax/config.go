package kafkax

import "time"

type Config struct {
	Brokers []string
	GroupID string

	// options
	MinBytes         int
	MaxBytes         int
	MaxWait          time.Duration
	CommitInterval   time.Duration
	StartOffset      int64 // kafka.FirstOffset hoặc kafka.LastOffset
	SessionTimeout   time.Duration
	RebalanceTimeout time.Duration
}

func (c *Config) clone() *Config {
	clone := *c
	return &clone
}
