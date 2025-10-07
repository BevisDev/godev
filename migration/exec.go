package migration

import "context"

type Exec interface {
	// Init sets up the migration system.
	// Returns an error if initialization fails.
	Init() error

	// Status retrieves the current migration status.
	// Returns an error if the status cannot be determined.
	Status() error

	// Up applies migrations up to the given version.
	// If version is 0, it may apply all pending migrations.
	// The context can be used to cancel or timeout the operation.
	// Returns an error if any migration fails.
	Up(c context.Context, version int64) error

	// Down rolls back migrations down to the given version.
	//
	// If version is 0, it may rollback all applied migrations.
	// The context can be used to cancel or timeout the operation.
	// Returns an error if rollback fails.
	Down(c context.Context, version int64) error
}
