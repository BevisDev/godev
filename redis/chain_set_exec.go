package redis

import "context"

type ChainSetExec[T any] interface {
	// Key specifies a single key to operate on for the next execution command
	Key(k string) ChainSetExec[T]

	// Values specifies multiple values to be stored with the key
	Values(vals interface{}) ChainSetExec[T]

	// Expire sets the Time-To-Live (TTL) for the key.
	// n is the time duration, and unit specifies the scale (e.g., "s" for seconds).
	Expire(n int, unit string) ChainSetExec[T]

	// Add adds one or more members to the set
	Add(ctx context.Context) error

	// Remove removes one or more members from the set
	Remove(ctx context.Context) error

	// Contains checks if a value exists in the set
	Contains(ctx context.Context, val interface{}) (bool, error)

	// GetAll returns all members of the set
	GetAll(ctx context.Context) ([]*T, error)

	// Size returns the number of elements in the set
	Size(ctx context.Context) (int64, error)

	// Delete removes the specified key from Redis.
	Delete(ctx context.Context) error
}
