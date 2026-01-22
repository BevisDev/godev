package redis

import "errors"

var (
	// ErrMissingKey is returned when a key is required but not provided.
	ErrMissingKey = errors.New("use Key() before")

	// ErrMissingKeys is returned when keys are required but not provided.
	ErrMissingKeys = errors.New("use Keys() before")

	// ErrMissingPrefix is returned when a prefix is required but not provided.
	ErrMissingPrefix = errors.New("use Prefix() before")

	// ErrMissingValue is returned when a value is required but not provided.
	ErrMissingValue = errors.New("use Value() before")

	// ErrMissingValues is returned when values are required but not provided.
	ErrMissingValues = errors.New("use Values() before")

	// ErrMissingChannel is returned when a channel is required but not provided.
	ErrMissingChannel = errors.New("use Channel() before")

	// ErrMissingPushOrBatch is returned when batch data is required but not provided.
	ErrMissingPushOrBatch = errors.New("use Push() or Batch() before")
)
