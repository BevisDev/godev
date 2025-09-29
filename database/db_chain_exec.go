package database

import "context"

type DBChainExec[T any] interface {
	// Select specifies the columns to retrieve.
	Select(cols ...string) DBChainExec[T]

	// From sets the table to query.
	From(table string) DBChainExec[T]

	// Where adds a WHERE condition with optional args.
	Where(cond string, args ...interface{}) DBChainExec[T]

	// OrderBy sets the ORDER BY clause.
	OrderBy(order string) DBChainExec[T]

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
	Insert(columns []string, values []interface{}) DBChainExec[T]

	// Update builds and executes an UPDATE statement with given column-value pairs.
	// Requires a WHERE clause to target rows, otherwise all rows will be updated.
	// Update(values map[string]interface{}) DBChainExec[T]
}
