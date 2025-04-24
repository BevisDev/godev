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
	showQuery  bool
	TimeoutSec int
	kindDB     types.KindDB
}

func NewDB(cf *ConfigDB) (*Database, error) {
	if cf.TimeoutSec <= 0 {
		cf.TimeoutSec = defaultTimeoutSec
	}
	database := &Database{
		showQuery:  cf.ShowQuery,
		TimeoutSec: cf.TimeoutSec,
		kindDB:     cf.Kind,
	}
	db, err := database.newConnection(cf)
	database.DB = db
	return database, err
}

func (d *Database) newConnection(cf *ConfigDB) (*sqlx.DB, error) {
	var (
		connStr string
		db      *sqlx.DB
		err     error
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
		return nil, errors.New("unsupported database kind " + string(cf.Kind))
	}

	// connect
	db, err = sqlx.Connect(types.SQLDriver[cf.Kind].String(), connStr)
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

func (d *Database) viewQuery(query string) {
	if d.showQuery {
		log.Printf("Query: %s\n", query)
	}
}

func (d *Database) GetParam() string {
	switch d.kindDB {
	case types.SqlServer:
		return "@p"
	case types.Postgres:
		return "$"
	default: // mysql
		return "?"
	}
}

func (d *Database) BeginTrans(ctx context.Context) (*sqlx.Tx, error) {
	return d.DB.BeginTxx(ctx, nil)
}

func (d *Database) mustBePtr(dest interface{}) (err error) {
	if !validate.IsPtr(dest) {
		return errors.New("must be a pointer")
	}
	return
}

func (d *Database) IsNoResult(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (d *Database) isIn(query string) bool {
	return strings.Contains(query, "IN") || strings.Contains(query, "in")
}

func (d *Database) GetList(c context.Context, dest interface{}, query string, args ...interface{}) (err error) {
	err = d.mustBePtr(dest)
	if err != nil {
		return
	}

	if d.isIn(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return
		}
	}
	query = d.DB.Rebind(query)
	d.viewQuery(query)

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	if validate.IsNilOrEmpty(args) {
		return d.DB.SelectContext(ctx, dest, query)
	}
	return d.DB.SelectContext(ctx, dest, query, args...)
}

func (d *Database) GetAny(c context.Context, dest interface{}, query string, args ...interface{}) (err error) {
	err = d.mustBePtr(dest)
	if err != nil {
		return err
	}

	if d.isIn(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return
		}
	}
	query = d.DB.Rebind(query)
	d.viewQuery(query)

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	if validate.IsNilOrEmpty(args) {
		return d.DB.GetContext(ctx, dest, query)
	}
	return d.DB.GetContext(ctx, dest, query, args...)
}

func (d *Database) ExecQuery(c context.Context, query string, args ...interface{}) (err error) {
	if d.isIn(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return
		}
	}
	query = d.DB.Rebind(query)
	d.viewQuery(query)
	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.BeginTrans(ctx)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return
	}
	return
}

// Save using for Insert Or Update query
// Any named placeholder parameters are replaced with fields from arg.
// Example query: INSERT INTO person (first_name,last_name,email) VALUES (:first,:last,:email)
// args: map[string]interface{}{ "first": "Bin","last": "Smuth", "email": "bensmith@allblacks.nz"}
// or struct with the `db` tag
func (d *Database) Save(c context.Context, query string, args interface{}) (err error) {
	d.viewQuery(query)
	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.BeginTrans(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.NamedExecContext(ctx, query, args)
	if err != nil {
		return
	}
	return
}

// InsertedId inserts record and returns id
// LastInsertId function should not be used with this SQL Server driver
// Please use OUTPUT clause or SCOPE_IDENTITY() to the end of your query
func (d *Database) InsertedId(c context.Context, query string, args ...interface{}) (id int, err error) {
	d.viewQuery(query)
	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.BeginTrans(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = tx.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return
	}
	return
}

func (d *Database) formatRow(idx int) string {
	var p = d.GetParam()
	if types.MySQL == d.kindDB {
		return p
	}
	return fmt.Sprintf("%s%d", p, idx)
}

func (d *Database) InsertMany(c context.Context, query string, size, col int, args ...interface{}) error {
	if size <= 0 {
		return errors.New("size must be greater than 0")
	}
	var placeholders []string
	for i := 0; i < size; i++ {
		var row []string
		for j := 1; j <= col; j++ {
			row = append(row, d.formatRow(i*col+j))
		}
		placeholders = append(placeholders, "("+strings.Join(row, ", ")+")")
	}
	query += strings.Join(placeholders, ", ")
	return d.ExecQuery(c, query, args...)
}

func (d *Database) UpdateMany(c context.Context, query string, entities []interface{}) (err error) {
	d.viewQuery(query)
	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.BeginTrans(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	for _, e := range entities {
		_, err = tx.NamedExecContext(ctx, query, e)
		if err != nil {
			return
		}
	}
	return
}

func (d *Database) Delete(c context.Context, query string, args interface{}) (err error) {
	d.viewQuery(query)
	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.BeginTrans(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.NamedExecContext(ctx, query, args)
	if err != nil {
		return
	}
	return
}
