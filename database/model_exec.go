package database

import "context"

// ModelExec defines the CRUD operations for a model bound to a table.
type ModelExec[T any] interface {
	// Where adds a WHERE condition with optional args.
	Where(cond string, args ...interface{}) ModelExec[T]

	// First returns the first matching row or nil if none found.
	First(ctx context.Context) (*T, error)

	// Find returns all matching rows.
	Find(ctx context.Context) ([]*T, error)

	// Create inserts a new row based on data.
	// Returns the inserted row when supported by the database.
	Create(ctx context.Context, data any) (*T, error)

	// Updates updates matching rows with provided data.
	Updates(ctx context.Context, data any) (int64, error)

	// Delete deletes matching rows.
	Delete(ctx context.Context) (int64, error)

	// Count returns the number of matching rows.
	Count(ctx context.Context) (int64, error)
}
