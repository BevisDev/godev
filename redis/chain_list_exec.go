package redis

import "context"

type ChainListExec[T any] interface {
	// Key specifies a single key to operate on for the next execution command
	Key(k string) ChainListExec[T]

	// Values specifies multiple values to be stored with the key
	Values(vals interface{}) ChainListExec[T]

	// Expire sets the Time-To-Live (TTL) for the key.
	// n is the time duration, and unit specifies the scale (e.g., "s" for seconds).
	Expire(n int, unit string) ChainListExec[T]

	// AddFirst inserts one or more values at the head (left) of the list.
	AddFirst(ctx context.Context) error

	// Add inserts one or more values at the tail (right) of the list.
	Add(ctx context.Context) error

	// PopFront retrieves and removes the first element (head) of the list.
	PopFront(ctx context.Context) (*T, error)

	// Pop retrieves and removes the last element (tail) of the list.
	Pop(ctx context.Context) (*T, error)

	// GetRange returns a slice of elements between the specified start and stop indexes.
	GetRange(ctx context.Context) ([]*T, error)

	// Get retrieves the element at the specified index from the Redis list.
	Get(ctx context.Context, index int64) (*T, error)

	// Size returns the number of elements in the list.
	Size(ctx context.Context) (int64, error)

	// Delete removes the specified key from Redis.
	Delete(ctx context.Context) error
}
