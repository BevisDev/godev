package kafkax

import (
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// Config holds Kafka configuration
type Config struct {
	// Kafka brokers
	Brokers []string

	// Producer config
	Producer ProducerConfig

	// Consumer config
	Consumer ConsumerConfig
}

func (c *Config) clone() *Config {
	clone := *c
	return &clone
}

type ProducerConfig struct {
	// Performance tuning
	BatchSize    int
	BatchTimeout time.Duration
	MaxAttempts  int
	Compression  compress.Compression

	// Async or sync
	Async bool

	// Required acks (-1, 0, 1)
	// -1 = all replicas, 0 = no ack, 1 = leader only
	RequiredAcks int

	// Balancer strategy
	Balancer kafka.Balancer

	// Idempotent writes (exactly-once semantics)
	Idempotent bool
}

type ConsumerConfig struct {
	GroupID string   // Consumer group ID
	Topics  []string // Topics to subscribe

	// Offset management
	StartOffset int64 // kafka.FirstOffset or kafka.LastOffset

	// Performance tuning
	CommitInterval time.Duration // Commit frequency (default: 1s)
	MaxWait        time.Duration // Max wait for messages (default: 500ms)
	MinBytes       int           // Min bytes to fetch (default: 1)
	MaxBytes       int           // Max bytes to fetch (default: 10MB)

	// Commit strategy
	AutoCommit bool // Auto vs manual commit (default: false)

	// Rebalancing
	PartitionWatchInterval time.Duration // (default: 5s)
	SessionTimeout         time.Duration // (default: 10s)
	RebalanceTimeout       time.Duration // (default: 30s)
	HeartbeatInterval      time.Duration // (default: 3s)

	// Isolation level
	IsolationLevel kafka.IsolationLevel // ReadCommitted or ReadUncommitted
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate brokers
	if len(c.Brokers) == 0 {
		return ErrNoBrokers
	}

	// Validate producer config
	if err := c.validateProducerConfig(); err != nil {
		return fmt.Errorf("invalid producer config: %w", err)
	}

	// Validate consumer config (if consumer will be initialized)
	if c.Consumer.GroupID != "" || len(c.Consumer.Topics) > 0 {
		if err := c.validateConsumerConfig(); err != nil {
			return fmt.Errorf("invalid consumer config: %w", err)
		}
	}

	return nil
}

func (c *Config) validateProducerConfig() error {
	if c.Producer.BatchSize < 1 {
		return fmt.Errorf("batch size must be >= 1")
	}

	if c.Producer.MaxAttempts < 1 {
		return fmt.Errorf("max attempts must be >= 1")
	}

	if c.Producer.RequiredAcks < -1 || c.Producer.RequiredAcks > 1 {
		return fmt.Errorf("required acks must be -1, 0, or 1")
	}

	return nil
}

func (c *Config) validateConsumerConfig() error {
	if c.Consumer.GroupID == "" {
		return ErrNoGroupID
	}

	if len(c.Consumer.Topics) == 0 {
		return ErrNoTopics
	}

	if c.Consumer.MinBytes < 0 {
		return fmt.Errorf("min bytes must be >= 0")
	}

	if c.Consumer.MaxBytes < c.Consumer.MinBytes {
		return fmt.Errorf("max bytes must be >= min bytes")
	}

	return nil
}
