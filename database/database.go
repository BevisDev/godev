package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/types"
	"github.com/BevisDev/godev/utils/validate"
	"log"
	"strings"
	"time"

	"github.com/BevisDev/godev/utils"
	"github.com/jmoiron/sqlx"
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

const defaultTimeoutSec = 10

type Database struct {
	DB         *sqlx.DB
	TimeoutSec int
	showQuery  bool
	kindDB     types.KindDB
}

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
	case types.Postgres:
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
	case types.Oracle:
		connStr = fmt.Sprintf("%s/%s@%s:%d/%s",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
	case types.MySQL:
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
	default:
		return nil, errors.New("unsupported database kind " + cf.Kind.String())
	}

	// connect
	db, err = sqlx.Connect(cf.Kind.GetDriver(), connStr)
	if err != nil {
		return nil, err
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
		return nil, err
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

func (d *Database) IsNoResult(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (d *Database) MustBePtr(dest interface{}) (err error) {
	if !validate.IsPtr(dest) {
		return errors.New("must be a pointer")
	}
	return
}

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
			panic(p)
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

func (d *Database) Execute(ctx context.Context, query string, tx *sqlx.Tx, args ...interface{}) (err error) {
	d.ViewQuery(query)
	if tx == nil {
		_, err = d.DB.ExecContext(ctx, query, args...)
	} else {
		_, err = tx.ExecContext(ctx, query, args...)
	}
	return
}

func (d *Database) ExecuteTx(ctx context.Context, query string, args ...interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Execute(ctx, query, tx, args...)
	})
}

func (d *Database) ExecuteSafe(ctx context.Context, query string, args ...interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelSerializable, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Execute(ctx, query, tx, args...)
	})
}

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

func (d *Database) SaveGettingId(ctx context.Context, query string, tx *sqlx.Tx, args ...interface{}) (id int, err error) {
	d.ViewQuery(query)
	if tx == nil {
		err = d.DB.QueryRowContext(ctx, query, args...).Scan(&id)
	} else {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
	}
	return
}

// InsertedId executes an INSERT query within a transaction and returns the newly inserted record's ID.
//
// This method is intended for SQL Server, where `LastInsertId` is not supported by the driver.
// To retrieve the inserted ID, your query must include the `OUTPUT` clause or use `SCOPE_IDENTITY()`.
//
// Example:
//
//	query := `INSERT INTO users (name, email) VALUES (@p1, @p2);
//	          SELECT SCOPE_IDENTITY();`
//
//	args := []interface{}{
//	    "John",
//	    "john@example.com",
//	}
//
//	id, err := db.InsertedId(ctx, query, args...)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("New ID:", id)
//
// Parameters:
//   - ctx: context for timeout or cancellation.
//   - query: the INSERT SQL query string.
//   - args: variadic parameters to match the placeholders in the query.
//
// Returns:
//   - id: the auto-generated primary key returned by the query.
//   - err: any error encountered during execution.
func (d *Database) InsertedId(ctx context.Context, query string, args ...interface{}) (id int, err error) {
	err = d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		id, err = d.SaveGettingId(ctx, query, tx, args...)
		return err
	})
	return
}

// InsertMany inserts multiple rows into the given table using bulk INSERT.
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
//	err := db.InsertMany(ctx, "users", 2, colNames, args...)
func (d *Database) InsertMany(ctx context.Context, table string, size int, colNames []string, args ...interface{}) error {
	sizeCol := len(colNames)
	if sizeCol <= 0 {
		return errors.New("size must be greater than 0")
	}

	var placeholders []string
	for i := 0; i < size; i++ {
		var row []string
		for j := 1; j <= sizeCol; j++ {
			row = append(row, d.FormatRow(i*sizeCol+j))
		}
		placeholders = append(placeholders, "("+strings.Join(row, ", ")+")")
	}

	cols := strings.Join(colNames, ", ")
	placeholderStr := strings.Join(placeholders, ", ")
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, cols, placeholderStr)
	return d.ExecuteTx(ctx, query, args...)
}

func (d *Database) Delete(ctx context.Context, query string, args interface{}) (err error) {
	return d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		return d.Save(ctx, tx, query, args)
	})
}

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
