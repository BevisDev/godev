package database

import "context"

type ChainExec[T any] interface {
	// Select specifies the columns to retrieve.
	Select(cols ...string) ChainExec[T]

	// From sets the table to query.
	From(table string) ChainExec[T]

	// Where adds a WHERE condition with optional args.
	Where(cond string, args ...interface{}) ChainExec[T]

	Top(n int) ChainExec[T]

	Limit(n int) ChainExec[T]

	Offset(n int) ChainExec[T]

	// OrderBy sets the ORDER BY clause.
	OrderBy(order string) ChainExec[T]

	// First executes a query and scans a single result into dest.
	// If the query returns no rows, it returns an error (sql.ErrNoRows),
	// which you can check with IsNoResult(err).
	First(ctx context.Context) (*T, error)

	// FirstOrNil executes the query and returns the first result.
	// Returns nil if no record is found.
	FirstOrNil(ctx context.Context) (*T, error)

	// FindAll executes the query and returns all results as a slice.
	FindAll(ctx context.Context) ([]*T, error)

	// Insert builds an INSERT statement with given columns and values.
	Insert(ctx context.Context, data any) (*T, error)

	// Update builds and executes an UPDATE statement with given column-value pairs.
	// Requires a WHERE clause to target rows, otherwise all rows will be updated.
	// Update(values map[string]interface{}) ChainExec[T]
}
