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

type ConfigDB struct {
	Kind           types.KindDB
	Schema         string
	TimeoutSec     int
	Host           string
	Port           int
	Username       string
	Password       string
	MaxOpenConns   int
	MaxIdleConns   int
	MaxIdleTimeSec int
	MaxLifeTimeSec int
	ShowQuery      bool
	Params         map[string]string
}

var defaultTimeoutSec = 30

type Database struct {
	DB         *sqlx.DB
	TimeoutSec int
	showQuery  bool
	kindDB     types.KindDB
}

func NewDB(cf *ConfigDB) (*Database, error) {
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

// RunTx is template transaction with callback func
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

// SaveTx using for Insert Or Update query with Transaction
// Any named placeholder parameters are replaced with fields from arg.
// Example query: INSERT INTO person (first_name,last_name,email) VALUES (:first,:last,:email)
// args: map[string]interface{}{ "first": "Bin","last": "Smith", "email": "bensmith@allblacks.nz"}
// or struct with the `db` tag
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

// InsertedId inserts record and returns id
// LastInsertId function should not be used with this SQL Server driver
// Please use OUTPUT clause or SCOPE_IDENTITY() to the end of your query
func (d *Database) InsertedId(ctx context.Context, query string, args ...interface{}) (id int, err error) {
	err = d.RunTx(ctx, sql.LevelDefault, func(ctx context.Context, tx *sqlx.Tx) error {
		id, err = d.SaveGettingId(ctx, query, tx, args...)
		return err
	})
	return
}

func (d *Database) InsertMany(c context.Context, query string, size, col int, args ...interface{}) error {
	if size <= 0 {
		return errors.New("size must be greater than 0")
	}
	var placeholders []string
	for i := 0; i < size; i++ {
		var row []string
		for j := 1; j <= col; j++ {
			row = append(row, d.FormatRow(i*col+j))
		}
		placeholders = append(placeholders, "("+strings.Join(row, ", ")+")")
	}
	query += strings.Join(placeholders, ", ")
	return d.ExecuteTx(c, query, args...)
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
