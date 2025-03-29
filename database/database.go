package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/BevisDev/godev/utils"
	"github.com/jmoiron/sqlx"
)

type ConfigDB struct {
	Kind         string
	Schema       string
	TimeoutSec   int
	Host         string
	Port         int
	Username     string
	Password     string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifeTime  int
	ShowQuery    bool
}

type Database struct {
	DB         *sqlx.DB
	showQuery  bool
	TimeoutSec int
}

func NewDB(cf *ConfigDB) (*Database, error) {
	database := &Database{
		showQuery:  cf.ShowQuery,
		TimeoutSec: cf.TimeoutSec,
	}
	db, err := database.newConnection(cf)
	database.DB = db
	return database, err
}

func (d Database) newConnection(cf *ConfigDB) (*sqlx.DB, error) {
	var (
		connStr string
		db      *sqlx.DB
		err     error
		driver  string
	)

	// build connectionString
	switch cf.Kind {
	case utils.SQLServer:
		connStr = fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s",
			cf.Host, cf.Port, cf.Username, cf.Password, cf.Schema)
		driver = "sqlserver"
		break
	case utils.Postgres:
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cf.Host, cf.Port, cf.Username, cf.Password, cf.Schema)
		driver = "postgres"
	case utils.Oracle:
		connStr = fmt.Sprintf("user=%s password=%s connectString=%s:%d/%s",
			cf.Username, cf.Password, cf.Host, cf.Port, cf.Schema)
		driver = "godror"
		break
	default:
		return nil, errors.New("unsupported database kind " + cf.Kind)
	}

	// connect
	db, err = sqlx.Connect(driver, connStr)
	if err != nil {
		return nil, err
	}

	// set pool
	db.SetMaxOpenConns(cf.MaxOpenConns)
	db.SetMaxIdleConns(cf.MaxIdleConns)
	db.SetConnMaxIdleTime(time.Duration(cf.MaxLifeTime) * time.Second)

	// ping check connection
	if err = db.Ping(); err != nil {
		return nil, err
	}
	log.Printf("connect db %s success \n", cf.Schema)
	return db, nil
}

func (d Database) Close() {
	d.DB.Close()
}

func (d Database) viewQuery(query string) {
	if d.showQuery {
		log.Printf("Query: %s\n", query)
	}
}

func (d Database) BeginTrans() (*sqlx.Tx, error) {
	return d.DB.Beginx()
}

func (d Database) mustBePointer(dest interface{}) error {
	if !utils.IsPointer(dest) {
		return errors.New("must be a pointer")
	}
	return nil
}

func (d Database) isIn(query string) bool {
	return strings.Contains(query, "IN") || strings.Contains(query, "in")
}

func (d Database) GetList(c context.Context, dest interface{}, query string, args ...interface{}) error {
	err := d.mustBePointer(dest)
	if err != nil {
		return err
	}

	if d.isIn(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return err
		}
	}
	query = d.DB.Rebind(query)
	d.viewQuery(query)

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	if utils.IsNilOrEmpty(args) {
		return d.DB.SelectContext(ctx, dest, query)
	}

	return d.DB.SelectContext(ctx, dest, query, args...)
}

func (d Database) GetAny(c context.Context, dest interface{}, query string, args ...interface{}) error {
	err := d.mustBePointer(dest)
	if err != nil {
		return err
	}

	if d.isIn(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return err
		}
	}
	query = d.DB.Rebind(query)
	d.viewQuery(query)

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	if utils.IsNilOrEmpty(args) {
		return d.DB.GetContext(ctx, dest, query)
	}

	return d.DB.GetContext(ctx, dest, query, args...)
}

func (d Database) ExecQuery(c context.Context, query string, args ...interface{}) error {
	var err error

	if d.isIn(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return err
		}
	}
	query = d.DB.Rebind(query)
	d.viewQuery(query)

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = d.DB.ExecContext(ctx, query, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// Save using for Insert Or Update query
// Any named placeholder parameters are replaced with fields from arg.
// Example query: INSERT INTO person (first_name,last_name,email) VALUES (:first,:last,:email)
// args: map[string]interface{}{ "first": "Bin","last": "Smuth", "email": "bensmith@allblacks.nz"}
// or struct with the `db` tag
func (d Database) Save(c context.Context, query string, args interface{}) error {
	d.viewQuery(query)

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = d.DB.NamedExecContext(ctx, query, args)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return err
}

// InsertedId inserts record and returns id
// LastInsertId function should not be used with this SQL Server driver
// Please use OUTPUT clause or SCOPE_IDENTITY() to the end of your query
func (d Database) InsertedId(c context.Context, query string, args ...interface{}) (int, error) {
	var id int
	d.viewQuery(query)

	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.DB.BeginTxx(ctx, nil)
	if err != nil {
		return id, err
	}

	err = d.DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		tx.Rollback()
		return id, err
	}

	tx.Commit()
	return id, err
}

func (d Database) Delete(c context.Context, query string, args interface{}) error {
	d.viewQuery(query)
	ctx, cancel := utils.CreateCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	tx, err := d.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = d.DB.NamedExecContext(ctx, query, args)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return err
}
