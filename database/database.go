package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/validate"
	"github.com/jmoiron/sqlx"
)

const (
	// MaxParams defines the maximum number of parameters allowed
	// To avoid hitting this hard limit, it's recommended to stay under 2000.
	// This value is used to determine safe batch sizes for bulk operations
	maxParams = 2000
)

// DB represents a database connection along with configuration
// settings and options for query logging.
//
// It embeds *Config to provide access to database configuration,
// and maintains an internal sqlx.DB connection for executing queries.
type DB struct {
	*Config
	db *sqlx.DB // db is the initialized sqlx.DB connection.
}

// New creates a new DB instance from the given Config.
//
// It applies default values, initializes connection settings (pool, timeout),
// connects to the appropriate database based on DBType (e.g., SQL Server, Postgres),
// and performs a ping to verify connectivity.
func New(cfg *Config) (*DB, error) {
	if cfg == nil {
		return nil, errors.New("[database] config is nil")
	}

	// Apply defaults
	cfg.clone()

	db := &DB{Config: cfg}

	// Initialize connection
	dbx, err := db.connect()
	if err != nil {
		return nil, err
	}
	db.db = dbx

	return db, nil
}

// connect establishes a database connection using the configured settings.
func (d *DB) connect() (*sqlx.DB, error) {
	cfg := d.Config

	// Get connection string
	connStr := cfg.getDSN()
	if connStr == "" {
		return nil, fmt.Errorf("[database] unsupported database type: %s", cfg.DBType.String())
	}

	// Connect to database
	db, err := sqlx.Connect(cfg.DBType.GetDriver(), connStr)
	if err != nil {
		return nil, fmt.Errorf("[database] failed to connect: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.MaxIdleTime)
	db.SetConnMaxLifetime(cfg.MaxLifeTime)

	// Verify connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("[database] ping failed: %w", err)
	}

	log.Printf("[database] connected to %s successfully", cfg.DBName)
	return db, nil
}

// Ping verifies the database connection is still alive.
func (d *DB) Ping() error {
	if d.db == nil {
		return errors.New("[database] ping error")
	}
	return d.db.Ping()
}

// Close closes the database connection and releases resources.
func (d *DB) Close() {
	if d.db != nil {
		_ = d.db.Close()
		d.db = nil
	}
}

// GetDB returns the underlying sqlx.DB connection.
func (d *DB) GetDB() *sqlx.DB {
	return d.db
}

// ViewQuery logs the SQL query if ShowQuery is enabled.
func (d *DB) ViewQuery(query string) {
	if d.ShowQuery {
		log.Printf("[database] query: %s", query)
	}
}

// IsNoResult returns true if the error indicates no rows were found.
func (d *DB) IsNoResult(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// MustBePtr validates that dest is a non-nil pointer.
func (d *DB) MustBePtr(dest interface{}) error {
	if !validate.IsNonNilPointer(dest) {
		return errors.New("[database] destination must be a non-nil pointer")
	}
	return nil
}

// GetTemplate returns the SQL template for the given template type and database type.
func (d *DB) GetTemplate(template TemplateJSON) string {
	if tempDB, ok := TemplateDBMap[d.DBType]; ok {
		if tpl, ok := tempDB[template]; ok {
			return tpl
		}
	}
	return ""
}

// FormatRow formats a parameter placeholder for the current database type.
// For MySQL, returns "?"; for others, returns formatted placeholder with index (e.g., "$1", "@p1").
func (d *DB) FormatRow(idx int) string {
	placeholder := d.DBType.GetPlaceHolder()
	if d.DBType == MySQL {
		return placeholder
	}
	return fmt.Sprintf("%s%d", placeholder, idx)
}

// rebind processes the query string and arguments for database-specific placeholder binding.
// It handles IN clauses and rebinds placeholders according to the database type.
func (d *DB) rebind(query string, args ...interface{}) (string, []interface{}, error) {
	// Handle IN clauses (expand slice arguments)
	if strings.Contains(strings.ToUpper(query), "IN") {
		var err error
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return query, args, fmt.Errorf("[database] failed to expand IN clause: %w", err)
		}
	}

	// Rebind placeholders for current database type
	db := d.GetDB()
	query = db.Rebind(query)

	d.ViewQuery(query)
	return query, args, nil
}

// RunTx runs a function within a database transaction with the specified isolation level.
//
// It handles transaction lifecycle (begin, commit, rollback) and recovers from panics.
// If the function returns an error or panics, the transaction is rolled back.
func (d *DB) RunTx(ctx context.Context, level sql.IsolationLevel,
	fn func(ctx context.Context, tx *sqlx.Tx) error,
) error {
	txCtx, cancel := utils.NewCtxTimeout(ctx, d.Timeout)
	defer cancel()

	db := d.GetDB()
	tx, err := db.BeginTxx(txCtx, &sql.TxOptions{
		Isolation: level,
	})
	if err != nil {
		return fmt.Errorf("[database] failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("[database] panic recovered in transaction: %v\n%s", p, debug.Stack())
		}
		if err != nil {
			_ = tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("[database] failed to commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(txCtx, tx)
	return err
}

// GetList executes a query and scans all resulting rows into dest.
//
// dest must be a pointer to a slice of structs or values.
// If no rows are returned, dest will remain an empty slice (no error is thrown).
func (d *DB) GetList(c context.Context, dest interface{}, query string, args ...interface{}) error {
	if err := d.MustBePtr(dest); err != nil {
		return err
	}

	query, newArgs, err := d.rebind(query, args...)
	if err != nil {
		return err
	}

	ctx, cancel := utils.NewCtxTimeout(c, d.Timeout)
	defer cancel()

	db := d.GetDB()
	if validate.IsNilOrEmpty(newArgs) {
		return db.SelectContext(ctx, dest, query)
	}
	return db.SelectContext(ctx, dest, query, newArgs...)
}

// GetAny executes a query and scans a single result into dest.
//
// dest must be a pointer to a value or struct.
// If the query returns no rows, it returns an error (sql.ErrNoRows),
// which you can check with IsNoResult(err).
func (d *DB) GetAny(c context.Context, dest interface{}, query string, args ...interface{}) error {
	if err := d.MustBePtr(dest); err != nil {
		return err
	}

	query, newArgs, err := d.rebind(query, args...)
	if err != nil {
		return err
	}

	ctx, cancel := utils.NewCtxTimeout(c, d.Timeout)
	defer cancel()

	db := d.GetDB()
	if validate.IsNilOrEmpty(newArgs) {
		return db.GetContext(ctx, dest, query)
	}
	return db.GetContext(ctx, dest, query, newArgs...)
}

// Execute runs the given SQL query with optional arguments.
// If a transaction is provided, the query runs within it.
// Otherwise, it executes directly on the database connection.
func (d *DB) Execute(ctx context.Context, query string, tx *sqlx.Tx, args ...interface{}) error {
	d.ViewQuery(query)

	if tx != nil {
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}

	db := d.GetDB()
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// ExecuteTx runs the query in a new transaction with default isolation level.
// Rolls back if an error occurs.
func (d *DB) ExecuteTx(ctx context.Context, query string, args ...interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Execute(ctx, query, tx, args...)
	})
}

// ExecuteSafe runs the query in a new transaction with serializable isolation.
// Ensures maximum data safety.
func (d *DB) ExecuteSafe(ctx context.Context, query string, args ...interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelSerializable, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Execute(ctx, query, tx, args...)
	})
}

// ExecReturningId executes a query that returns a single auto-generated ID.
//
// Returns the generated ID and any error encountered.
func (d *DB) ExecReturningId(ctx context.Context, query string, args ...interface{}) (int, error) {
	d.ViewQuery(query)

	db := d.GetDB()
	var id int
	err := db.QueryRowxContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("[database] failed to get returned ID: %w", err)
	}
	return id, nil
}

// Prepare creates a prepared statement for later execution.
func (d *DB) Prepare(ctx context.Context, query string) (*sqlx.Stmt, error) {
	db := d.GetDB()
	stmt, err := db.PreparexContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("[database] failed to prepare statement: %w", err)
	}
	return stmt, nil
}

// Save executes a query with named parameters.
//
// The query should use named placeholders (e.g., :name).
// If tx is nil, the query is executed using the default connection;
// otherwise, it is executed within the provided transaction.
//
// Returns any error encountered during execution.
func (d *DB) Save(ctx context.Context, tx *sqlx.Tx, query string, args interface{}) (err error) {
	d.ViewQuery(query)

	if tx == nil {
		db := d.GetDB()
		_, err = db.NamedExecContext(ctx, query, args)
	} else {
		_, err = tx.NamedExecContext(ctx, query, args)
	}
	return
}

// InsertOrUpdate executes an SQL statement to insert a new record or update an existing one.
//
// It delegates the actual execution to the Save method, which handles building
// and executing the query with the provided arguments.
//
// Parameters:
//   - ctx: Context used to control cancellation and timeouts for the database operation.
//   - query: The SQL query string, typically containing parameter placeholders (e.g., `?` or `$1`).
//   - args: The arguments to bind to the query placeholders. Can be a struct, map, or slice,
//     depending on the underlying Save method implementation.
//
// Returns:
//   - error: An error if the database operation fails, otherwise nil.
func (d *DB) InsertOrUpdate(ctx context.Context, query string, args interface{}) error {
	return d.Save(ctx, nil, query, args)
}

// SaveTx executes an INSERT or UPDATE SQL statement within a transaction,
// using named placeholder parameters.
//
// Any named placeholders in the query (e.g., :first, :last) are replaced
// with values from `arg`, which can be either a `map[string]interface{}`
// or a struct tagged with `db`.
//
// Example:
//
//	query := `INSERT INTO person (first_name, last_name, email)
//	          VALUES (:first, :last, :email)`
//
//	args := map[string]interface{}{
//	    "first": "Bin",
//	    "last":  "Smith",
//	    "email": "bensmith@allblacks.nz",
//	}
//
//	// or use a struct:
//	type Person struct {
//	    First string `db:"first"`
//	    Last  string `db:"last"`
//	    Email string `db:"email"`
//	}
//	SaveTx(ctx, db, query, Person{...})
func (d *DB) SaveTx(ctx context.Context, query string, args interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Save(ctx, tx, query, args)
	})
}

// SaveSafe executes a query with named parameters within a transaction
// using serializable isolation level for maximum safety.
func (d *DB) SaveSafe(ctx context.Context, query string, args interface{}) error {
	return d.RunTx(ctx, sql.LevelSerializable, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Save(ctx, tx, query, args)
	})
}

// InsertReturning executes an INSERT query and scans the result into dest.
//
// Usage:
//   - If dest is *int: the query must return a single ID column.
//   - If dest is a struct pointer: the query must return all the columns
//     (e.g., using `OUTPUT INSERTED.*` or `RETURNING *`).
//
// DB-specific notes:
//   - SQL Server:
//   - Use "OUTPUT INSERTED.id" for ID only.
//   - Use "OUTPUT INSERTED.*" for the full inserted row.
//   - PostgreSQL:
//   - Use "RETURNING id" for ID only.
//   - Use "RETURNING *" for the full inserted row.
//   - MySQL:
//   - Does not support OUTPUT or RETURNING directly.
//   - To get the last inserted ID, run "SELECT LAST_INSERT_ID()" after insert.
//   - For full row fetch, you must re-query manually using the returned ID.
//
// The function automatically determines whether to use `Scan` (for int)
// or `StructScan` (for structs) based on the type of dest.
func (d *DB) InsertReturning(c context.Context, query string, dest interface{}, args ...interface{}) error {
	if err := d.MustBePtr(dest); err != nil {
		return err
	}
	d.ViewQuery(query)

	ctx, cancel := utils.NewCtxTimeout(c, d.Timeout)
	defer cancel()

	db := d.GetDB()
	row := db.QueryRowxContext(ctx, query, args...)

	switch dest.(type) {
	case *int, *int64:
		return row.Scan(dest)
	default:
		return row.StructScan(dest)
	}
}

// InsertBulk inserts multiple rows into the given table using bulk INSERT.
//
// It builds placeholders dynamically and executes the insert in a single query.
// Automatically batches large inserts to avoid parameter limits.
//
// Params:
//   - table:     Table name.
//   - row:      Number of rows to insert.
//   - colNames:  Columns to insert (must match args order).
//   - args:      values to insert
//
// Example:
//
//	colNames := []string{"name", "email"}
//	args := []interface{}{"Alice", "alice@example.com", "Bob", "bob@example.com"}
//	err := db.InsertBulk(ctx, "users", 2, colNames, args...)
func (d *DB) InsertBulk(ctx context.Context, table string, row int, colNames []string, args ...interface{}) error {
	col := len(colNames)
	if col <= 0 {
		return errors.New("[database] column count must be greater than 0")
	}

	if len(args) != row*col {
		return fmt.Errorf("[database] expected %d arguments for %d rows with %d columns, got %d", row*col, row, col, len(args))
	}

	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		if len(args) > maxParams {
			batchRow := maxParams / col
			for start := 0; start < row; start += batchRow {
				end := start + batchRow
				if end > row {
					end = row
				}

				batchArgs := args[start*col : end*col]
				if err := d.InsertBatch(ctx, tx, table, colNames, col, end-start, batchArgs); err != nil {
					return err
				}
			}
		} else {
			if err := d.InsertBatch(ctx, tx, table, colNames, col, row, args); err != nil {
				return err
			}
		}
		return nil
	})
}

// InsertBatch executes a single batch INSERT statement with the provided arguments.
// This is a helper method used by InsertBulk for batching large inserts.
func (d *DB) InsertBatch(ctx context.Context, tx *sqlx.Tx,
	table string, colNames []string,
	col int, row int, args []interface{},
) error {
	// Build a placeholder row for each record
	// Example for MSSQL: (@p1, @p2), (@p3, @p4), ...
	var placeholders []string
	for i := 0; i < row; i++ {
		var paramRow []string
		for j := 1; j <= col; j++ {
			paramRow = append(paramRow, d.FormatRow(i*col+j))
		}
		// Add this row's placeholders to the list
		placeholders = append(placeholders, "("+strings.Join(paramRow, ", ")+")")
	}

	// Join column names into comma-separated string
	// Example: "name, email, age"
	colsJoin := strings.Join(colNames, ", ")

	// Join all row placeholder groups into final VALUES string
	// Example: "(?, ?), (?, ?), (?, ?)"
	placeholderStr := strings.Join(placeholders, ", ")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, colsJoin, placeholderStr)

	return d.Execute(ctx, query, tx, args...)
}

// InsertMany inserts multiple records into the database using a parameterized
// INSERT query and NamedExecContext for each entity.
// The entire operation is wrapped in a transaction at default isolation level.
// If any insert fails, the transaction is rolled back.
//
// Example:
//
//	query := `INSERT INTO users (name, email) VALUES (:name, :email)`
//	entities := []interface{}{
//	    map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
//	    map[string]interface{}{"name": "Bob", "email": "bob@example.com"},
//	}
//	err := db.InsertMany(ctx, query, entities)
//
// Note: The fields in each entity must match the named parameters in the query.
func (d *DB) InsertMany(ctx context.Context, query string, entities []interface{}) error {
	const batchSize = 1000
	d.ViewQuery(query)

	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		for i := 0; i < len(entities); i += batchSize {
			end := i + batchSize
			if end > len(entities) {
				end = len(entities)
			}

			batch := entities[i:end]
			for _, e := range batch {
				_, err := tx.NamedExecContext(ctx, query, e)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Delete runs a delete query within a transaction using default isolation level.
//
// The query should use named parameters matching the fields in args.
func (d *DB) Delete(ctx context.Context, query string, args interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Save(ctx, tx, query, args)
	})
}

// UpdateMany executes the same update query for multiple entities,
// each with its own named parameters, inside a single transaction.
//
// Uses default isolation level.
func (d *DB) UpdateMany(ctx context.Context, query string, entities []interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		d.ViewQuery(query)
		for _, e := range entities {
			_, err = tx.NamedExecContext(ctx, query, e)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateManySafe is like UpdateMany but runs with serializable isolation level,
// ensuring maximum safety in concurrent environments.
func (d *DB) UpdateManySafe(ctx context.Context, query string, entities []interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelSerializable, func(ctx context.Context, tx *sqlx.Tx) error {
		d.ViewQuery(query)
		for _, e := range entities {
			_, err = tx.NamedExecContext(ctx, query, e)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
