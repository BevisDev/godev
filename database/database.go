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

// Database represents a database connection along with configuration
// settings and options for query logging.
//
// Fields:
//   - DB: the sqlx.DB connection instance.
//   - TimeoutSec: default query timeout in seconds.
//   - showQuery: if true, executed SQL queries will be logged.
//   - kindDB: the type of database (e.g., mysql, postgres, sqlserver).
type Database struct {
	*Config
	db *sqlx.DB // DB is the initialized sqlx.DB connection.
}

// New creates a new `*Database` instance from the given ConfigDB.
//
// It initializes connection settings (pool, timeout), connects to the
// appropriate database based on the `Kind` (e.g., SQL Server, Postgres),
// and performs a ping to verify connectivity.
func New(cf *Config) (*Database, error) {
	if cf == nil {
		return nil, errors.New("config is nil")
	}

	var db = &Database{Config: cf}

	// initialize connection
	dbx, err := db.connect()
	if err != nil {
		return nil, err
	}
	db.db = dbx

	return db, err
}

func (d *Database) connect() (*sqlx.DB, error) {
	var (
		db  *sqlx.DB
		err error
		cf  = d.Config
	)
	// get connection string
	connStr := cf.getDSN()
	if connStr == "" {
		return nil, errors.New("[database] unsupported type " + cf.DBType.String())
	}

	// connect
	db, err = sqlx.Connect(cf.DBType.GetDriver(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// set pool
	db.SetMaxOpenConns(cf.MaxOpenConns)
	db.SetMaxIdleConns(cf.MaxIdleConns)
	db.SetConnMaxIdleTime(cf.MaxIdleTimeSec)
	db.SetConnMaxLifetime(cf.MaxLifeTimeSec)

	// ping check connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database failed: %w", err)
	}

	log.Printf("connect db %s success \n", cf.DBName)
	return db, nil
}

func (d *Database) Ping() error {
	db := d.GetDB()
	return db.Ping()
}

func (d *Database) Close() {
	if d.db != nil {
		_ = d.db.Close()
		d.db = nil
	}
}

func (d *Database) GetDB() *sqlx.DB {
	return d.db
}

func (d *Database) ViewQuery(query string) {
	if d.ShowQuery {
		log.Printf("Query: %s\n", query)
	}
}

// IsNoResult returns true if the error indicates no rows were found.
func (d *Database) IsNoResult(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (d *Database) MustBePtr(dest interface{}) (err error) {
	if !validate.IsNonNilPointer(dest) {
		return errors.New("must be a pointer")
	}
	return
}

func (d *Database) GetTemplate(template TemplateJSON) string {
	if tempDB, ok := TemplateDBMap[d.DBType]; ok {
		if tpl, ok := tempDB[template]; ok {
			return tpl
		}
	}
	return ""
}

func (d *Database) FormatRow(idx int) string {
	var p = d.DBType.GetPlaceHolder()
	if MySQL == d.DBType {
		return p
	}
	return fmt.Sprintf("%s%d", p, idx)
}

func (d *Database) rebind(query string, args ...interface{}) (string, []interface{}, error) {
	var err error

	// if query using IN
	if strings.Contains(strings.ToUpper(query), "IN") {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return query, args, err
		}
	}

	db := d.GetDB()
	query = db.Rebind(query)

	d.ViewQuery(query)
	return query, args, err
}

func (d *Database) RunTx(c context.Context, level sql.IsolationLevel,
	fn func(ctx context.Context, tx *sqlx.Tx) error,
) error {
	ctx, cancel := utils.NewCtxTimeout(c, d.Timeout)
	defer cancel()

	db := d.GetDB()
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: level,
	})
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("panic recovered in transaction: %v\n%s", p, debug.Stack())
		}
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(ctx, tx)
	return err
}

// GetList executes a query and scans all resulting rows into dest.
//
// dest must be a pointer to a slice of structs or values.
// If no rows are returned, dest will remain an empty slice (no error is thrown).
func (d *Database) GetList(c context.Context, dest interface{}, query string, args ...interface{}) error {
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
func (d *Database) GetAny(c context.Context, dest interface{}, query string, args ...interface{}) error {
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
// Otherwise, it executes directly on the database.
func (d *Database) Execute(ctx context.Context, query string, tx *sqlx.Tx, args ...interface{}) error {
	d.ViewQuery(query)
	var err error

	if tx == nil {
		db := d.GetDB()
		_, err = db.ExecContext(ctx, query, args...)
	} else {
		_, err = tx.ExecContext(ctx, query, args...)
	}

	return err
}

// ExecuteTx runs the query in a new transaction with default isolation level.
// Rolls back if an error occurs.
func (d *Database) ExecuteTx(ctx context.Context, query string, args ...interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Execute(ctx, query, tx, args...)
	})
}

// ExecuteSafe runs the query in a new transaction with serializable isolation.
// Ensures maximum data safety.
func (d *Database) ExecuteSafe(ctx context.Context, query string, args ...interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelSerializable, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Execute(ctx, query, tx, args...)
	})
}

// ExecReturningId executes a query
// that returns a single auto-generated ID.
//
// If tx is nil, the query is executed using the default connection;
// otherwise, it runs within the provided transaction.
//
// Returns the generated ID and any error encountered.
func (d *Database) ExecReturningId(ctx context.Context, query string, args ...interface{}) (id int, err error) {
	d.ViewQuery(query)

	db := d.GetDB()
	err = db.QueryRowxContext(ctx, query, args...).Scan(&id)
	return
}

func (d *Database) Prepare(ctx context.Context, query string) (*sqlx.Stmt, error) {
	db := d.GetDB()
	return db.PreparexContext(ctx, query)
}

// Save executes a query with named parameters.
//
// The query should use named placeholders (e.g., :name).
// If tx is nil, the query is executed using the default connection;
// otherwise, it is executed within the provided transaction.
//
// Returns any error encountered during execution.
func (d *Database) Save(ctx context.Context, tx *sqlx.Tx, query string, args interface{}) (err error) {
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
func (d *Database) InsertOrUpdate(ctx context.Context, query string, args interface{}) error {
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
func (d *Database) SaveTx(ctx context.Context, query string, args interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Save(ctx, tx, query, args)
	})
}

func (d *Database) SaveSafe(ctx context.Context, query string, args interface{}) (err error) {
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
// Database-specific notes:
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
func (d *Database) InsertReturning(c context.Context, query string, dest interface{}, args ...interface{}) error {
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
func (d *Database) InsertBulk(ctx context.Context, table string, row int, colNames []string, args ...interface{}) error {
	col := len(colNames)
	if col <= 0 {
		return errors.New("size column must be greater than 0")
	}

	if len(args) != row*col {
		return fmt.Errorf("expected %d arguments, got %d", row*col, len(args))
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

func (d *Database) InsertBatch(ctx context.Context, tx *sqlx.Tx,
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
func (d *Database) InsertMany(ctx context.Context, query string, entities []interface{}) error {
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
func (d *Database) Delete(ctx context.Context, query string, args interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Save(ctx, tx, query, args)
	})
}

// UpdateMany executes the same update query for multiple entities,
// each with its own named parameters, inside a single transaction.
//
// Uses default isolation level.
func (d *Database) UpdateMany(ctx context.Context, query string, entities []interface{}) (err error) {
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
func (d *Database) UpdateManySafe(ctx context.Context, query string, entities []interface{}) (err error) {
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
