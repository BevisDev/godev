package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/types"
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/validate"
	"github.com/jmoiron/sqlx"
	"log"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

// ConfigDB defines the configuration for connecting to a SQL database.
//
// It supports common settings such as connection parameters, pool sizing, timeouts,
// and connection string customization. The `Kind` field determines the target
// database type (e.g., MySQL, Postgres, SQL Server, Oracle).
type ConfigDB struct {
	// Kind specifies the type of database (e.g., types.MySQL, types.Postgres).
	Kind types.KindDB

	// Schema is the name of the target database/schema.
	Schema string

	// TimeoutSec defines the default timeout (in seconds) for DB operations.
	TimeoutSec int

	// Host is the hostname or IP address of the database server.
	Host string

	// Port is the port number to connect to the database server.
	Port int

	// Username is the database login username.
	Username string

	// Password is the database login password.
	Password string

	// MaxOpenConns sets the maximum number of open connections to the database.
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections in the pool.
	MaxIdleConns int

	// MaxIdleTimeSec is the maximum amount of time (in seconds) a connection can remain idle.
	MaxIdleTimeSec int

	// MaxLifeTimeSec is the maximum amount of time (in seconds) a connection can be reused.
	MaxLifeTimeSec int

	// ShowQuery enables SQL query logging when set to true.
	ShowQuery bool

	// Params is an optional map of additional connection string parameters.
	Params map[string]string
}

// Database represents a database connection along with configuration
// settings and options for query logging.
//
// Fields:
//   - DB: the sqlx.DB connection instance.
//   - TimeoutSec: default query timeout in seconds.
//   - showQuery: if true, executed SQL queries will be logged.
//   - kindDB: the type of database (e.g., mysql, postgres, sqlserver).
type Database struct {
	// DB is the initialized sqlx.DB connection.
	DB *sqlx.DB

	// TimeoutSec specifies the default timeout for queries, in seconds.
	TimeoutSec int

	// showQuery enables logging of SQL queries when set to true.
	showQuery bool

	// kindDB stores the database type.
	// For example: sqlserver, postgres, mysql.
	kindDB types.KindDB
}

const (
	// defaultTimeoutSec defines the default timeout (in seconds) for database operations.
	defaultTimeoutSec = 30

	// MaxParams defines the maximum number of parameters allowed
	// To avoid hitting this hard limit, it's recommended to stay under 2000.
	// This value is used to determine safe batch sizes for bulk operations
	MaxParams = 2000
)

// NewDB creates a new `*Database` instance from the given ConfigDB.
//
// It initializes connection settings (pool, timeout), connects to the
// appropriate database based on the `Kind` (e.g., SQL Server, Postgres),
// and performs a ping to verify connectivity.
//
// If the config is nil, or connection fails, it returns an error.
//
// Example:
//
//	db, err := NewDB(&ConfigDB{
//	    Kind:     types.Postgres,
//	    Host:     "localhost",
//	    Port:     5432,
//	    Username: "admin",
//	    Password: "secret",
//	    Schema:   "mydb",
//	})
func NewDB(cf *ConfigDB) (*Database, error) {
	if cf == nil {
		return nil, errors.New("config is nil")
	}
	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}

	db, err := newConnection(cf)
	if err != nil {
		return nil, err
	}

	return &Database{
		showQuery:  cf.ShowQuery,
		TimeoutSec: cf.TimeoutSec,
		kindDB:     cf.Kind,
		DB:         db,
	}, err
}

// newConnection establishes a new `sqlx.DB` connection based on the provided ConfigDB.
//
// It builds the connection string depending on the database type (Kind),
// sets connection pool parameters if configured, and verifies the connection with a ping.
//
// Supported databases: SQL Server, Postgres, Oracle, MySQL.
//
// Returns an error if the database kind is unsupported or connection fails.
func newConnection(cf *ConfigDB) (*sqlx.DB, error) {
	var (
		db      *sqlx.DB
		err     error
		connStr string
	)
	// build connectionString
	switch cf.Kind {
	case types.SqlServer:
		connStr = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
		if len(cf.Params) > 0 {
			params := url.Values{}
			for k, v := range cf.Params {
				params.Add(k, v)
			}
			connStr += "&" + params.Encode()
		}
	case types.Postgres:
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
		if len(cf.Params) > 0 {
			params := url.Values{}
			for k, v := range cf.Params {
				params.Add(k, v)
			}
			connStr += "&" + params.Encode()
		}
	case types.Oracle:
		connStr = fmt.Sprintf("%s/%s@%s:%d/%s",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
	case types.MySQL:
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
		if len(cf.Params) > 0 {
			params := url.Values{}
			for k, v := range cf.Params {
				params.Add(k, v)
			}
			connStr += "?" + params.Encode()
		}
	default:
		return nil, errors.New("unsupported database kind " + cf.Kind.String())
	}

	// connect
	db, err = sqlx.Connect(cf.Kind.GetDriver(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// set pool
	if cf.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cf.MaxOpenConns)
	}
	if cf.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cf.MaxIdleConns)
	}
	if cf.MaxIdleTimeSec > 0 {
		db.SetConnMaxIdleTime(time.Duration(cf.MaxIdleTimeSec) * time.Second)
	}
	if cf.MaxLifeTimeSec > 0 {
		db.SetConnMaxLifetime(time.Duration(cf.MaxLifeTimeSec) * time.Second)
	}

	// ping check connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database failed: %w", err)
	}

	log.Printf("connect db %s success \n", cf.Schema)
	return db, nil
}

func (d *Database) Close() {
	d.DB.Close()
}

func (d *Database) ViewQuery(query string) {
	if d.showQuery {
		log.Printf("Query: %s\n", query)
	}
}

// IsNoResult returns true if the error indicates no rows were found.
func (d *Database) IsNoResult(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (d *Database) MustBePtr(dest interface{}) (err error) {
	if !validate.IsPtr(dest) {
		return errors.New("must be a pointer")
	}
	return
}

// GetTemplate returns a database-specific SQL template string for rendering JSON output,
// based on the configured database kind (e.g., MySQL, MSSQL, PostgreSQL) and the desired JSON format.
//
// Supported templates include:
//   - types.TemplateJSONArray: for returning JSON arrays (e.g., list of rows)
//   - types.TemplateJSONObject: for returning a single JSON object (e.g., one row)
//
// Note:
//   - MySQL templates require three fmt.Sprintf() parameters: (columns, table, where)
//     Example: fmt.Sprintf(template, "'id', id", "users", "WHERE active = 1")
//   - MSSQL and Postgres templates wrap a full SELECT query inside, so you typically pass only one parameter.
//
// Example usage:
//
//	tpl := db.GetTemplate(types.TemplateJSONArray)
//	query := fmt.Sprintf(tpl, "SELECT id, name FROM users")
//
//	// For MySQL:
//	tpl := db.GetTemplate(types.TemplateJSONArray)
//	query := fmt.Sprintf(tpl, "'id', id, 'name', name", "users", "WHERE active = 1")
func (d *Database) GetTemplate(template types.DBJSONTemplate) string {
	switch d.kindDB {
	case types.SqlServer:
		if types.TemplateJSONArray == template {
			return types.MSSQLJSONArrayTemplate
		}
		return types.MSSQLJSONObjectTemplate

		// In MySQL, the JSON templates require explicit table and WHERE clause placeholders.
		// You must use fmt.Sprintf(template, columns, table, where) when applying this.
		// Unlike MSSQL or Postgres, which embed the full SELECT inside the template,
		// MySQL templates are split and require manual composition.
	case types.MySQL:
		if types.TemplateJSONArray == template {
			return types.MySQLJSONArrayTemplate
		}
		return types.MySQLJSONObjectTemplate

	case types.Postgres:
		if types.TemplateJSONArray == template {
			return types.PostgresJSONArrayTemplate
		}
		return types.PostgresJSONObjectTemplate

	default:
		return ""
	}
}

// RebindQuery prepares a SQL query and its arguments for execution,
// applying the correct placeholder format based on the database dialect.
//
// - For queries containing an IN clause, it uses sqlx.In() to expand slice arguments.
// - Then, it's rebinding the query using d.DB.Rebind() to match the database's placeholder style.
//
// Example:
//
//	Input query: "SELECT * FROM users WHERE id IN (?) AND status = ?"
//	Args:        []interface{}{[]int{1, 2, 3}, "active"}
//	Output:      "SELECT * FROM users WHERE id IN ($1,$2,$3) AND status = $4" (for Postgres)
//
// ViewQuery is called to optionally log or inspect the final query.
//
// Returns:
//   - The final bound query string
//   - The updated argument slice
//   - Any error from sqlx.In (if applicable)
func (d *Database) RebindQuery(query string, args ...interface{}) (string, []interface{}, error) {
	var err error

	if strings.Contains(strings.ToUpper(query), "IN") {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return query, args, err
		}
	}
	query = d.DB.Rebind(query)

	d.ViewQuery(query)
	return query, args, err
}

// FormatRow returns a database-specific placeholder for a given parameter index,
// used when dynamically constructing bulk insert queries or custom SQL.
//
// Behavior:
//   - For MySQL: always returns "?" (since MySQL uses anonymous placeholders)
//   - For Postgres: returns "$1", "$2", etc.
//   - For MSSQL: returns "@p1", "@p2", etc.
//
// This allows dynamically building row placeholders such as:
//
//	"(?, ?, ?)" or "($1, $2, $3)" depending on the database in use.
//
// Example:
//
//	fmt.Sprintf("INSERT INTO users (id, name) VALUES (%s)", d.FormatRow(1))
func (d *Database) FormatRow(idx int) string {
	var p = d.kindDB.GetPlaceHolder()
	if types.MySQL == d.kindDB {
		return p
	}
	return fmt.Sprintf("%s%d", p, idx)
}

// RunTx executes a function within a database transaction with a given isolation level.
//
// It automatically manages context timeout (`d.TimeoutSec`), begins a transaction,
// and ensures proper commit or rollback based on the outcome of the callback function.
//
// If the callback returns an error or a panic occurs, the transaction is rolled back.
// If the callback succeeds, the transaction is committed.
//
// Example:
//
//	err := db.RunTx(ctx, sql.LevelSerializable, func(txCtx context.Context, tx *sqlx.Tx) error {
//	    _, err := tx.ExecContext(txCtx, "UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
//	    if err != nil {
//	        return err
//	    }
//	    _, err = tx.ExecContext(txCtx, "UPDATE accounts SET balance = balance + 100 WHERE id = ?", 2)
//	    return err
//	})
//
//	if err != nil {
//	    log.Fatalf("transaction failed: %v", err)
//	}
func (d *Database) RunTx(c context.Context, level sql.IsolationLevel, fn func(ctx context.Context, tx *sqlx.Tx) error) (err error) {
	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.DB.BeginTxx(ctx, &sql.TxOptions{
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
	return
}

// GetList executes a query and scans all resulting rows into dest.
//
// dest must be a pointer to a slice of structs or values.
// If no rows are returned, dest will remain an empty slice (no error is thrown).
func (d *Database) GetList(c context.Context, dest interface{}, query string, args ...interface{}) error {
	if err := d.MustBePtr(dest); err != nil {
		return err
	}

	query, newArgs, err := d.RebindQuery(query, args...)
	if err != nil {
		return err
	}

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	if validate.IsNilOrEmpty(newArgs) {
		return d.DB.SelectContext(ctx, dest, query)
	}
	return d.DB.SelectContext(ctx, dest, query, newArgs...)
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

	query, newArgs, err := d.RebindQuery(query, args...)
	if err != nil {
		return err
	}

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	if validate.IsNilOrEmpty(newArgs) {
		return d.DB.GetContext(ctx, dest, query)
	}
	return d.DB.GetContext(ctx, dest, query, newArgs...)
}

// Execute runs the given SQL query with optional arguments.
// If a transaction is provided, the query runs within it.
// Otherwise, it executes directly on the database.
func (d *Database) Execute(ctx context.Context, query string, tx *sqlx.Tx, args ...interface{}) (err error) {
	d.ViewQuery(query)
	if tx == nil {
		_, err = d.DB.ExecContext(ctx, query, args...)
	} else {
		_, err = tx.ExecContext(ctx, query, args...)
	}
	return
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
	err = d.DB.QueryRowxContext(ctx, query, args...).Scan(&id)
	return
}

func (d *Database) Prepare(ctx context.Context, query string) (*sqlx.Stmt, error) {
	return d.DB.PreparexContext(ctx, query)
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
		_, err = d.DB.NamedExecContext(ctx, query, args)
	} else {
		_, err = tx.NamedExecContext(ctx, query, args)
	}
	return
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

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	row := d.DB.QueryRowxContext(ctx, query, args...)

	switch dest.(type) {
	case *int:
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
//   - size:      Number of rows to insert.
//   - colNames:  Columns to insert (must match args order).
//   - args:      values to insert
//
// Example:
//
//	colNames := []string{"name", "email"}
//	args := []interface{}{"Alice", "alice@example.com", "Bob", "bob@example.com"}
//	err := db.InsertBulk(ctx, "users", 2, colNames, args...)
func (d *Database) InsertBulk(ctx context.Context, table string, size int, colNames []string, args ...interface{}) error {
	sizeCol := len(colNames)
	if sizeCol <= 0 {
		return errors.New("size column must be greater than 0")
	}

	if len(args) != size*sizeCol {
		return fmt.Errorf("expected %d arguments, got %d", size*sizeCol, len(args))
	}

	if len(args) > MaxParams {
		return errors.New("args exceed the maximum params")
	}

	var placeholders []string
	for i := 0; i < size; i++ {
		// Build a placeholder row for each record
		// Example for MSSQL: (@p1, @p2), (@p3, @p4), ...
		var row []string
		for j := 1; j <= sizeCol; j++ {
			row = append(row, d.FormatRow(i*sizeCol+j))
		}
		// Add this row's placeholders to the list
		placeholders = append(placeholders, "("+strings.Join(row, ", ")+")")
	}

	// Join column names into comma-separated string
	// Example: "name, email, age"
	cols := strings.Join(colNames, ", ")

	// Join all row placeholder groups into final VALUES string
	// Example: "(?, ?), (?, ?), (?, ?)"
	placeholderStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, cols, placeholderStr)
	return d.ExecuteTx(ctx, query, args...)
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
