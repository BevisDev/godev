package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/BevisDev/godev/helper"

	//_ "github.com/denisenkom/go-mssqldb"
	//_ "github.com/godror/godror"
	//_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

type ConfigDB struct {
	Profile      string
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
}

type Database struct {
	db         *sqlx.DB
	profile    string
	timeoutSec int
}

func NewDB(cf *ConfigDB) (*Database, error) {
	database := &Database{
		profile:    cf.Profile,
		timeoutSec: cf.TimeoutSec,
	}
	db, err := database.newConnection(cf)
	database.db = db
	return database, err
}

func (d *Database) newConnection(cf *ConfigDB) (*sqlx.DB, error) {
	var (
		connStr string
		db      *sqlx.DB
		err     error
		driver  string
	)

	// build connectionString
	switch cf.Kind {
	case helper.SQLServer:
		connStr = fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s",
			cf.Host, cf.Port, cf.Username, cf.Password, cf.Schema)
		driver = "sqlserver"
		break
	case helper.Postgres:
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cf.Host, cf.Port, cf.Username, cf.Password, cf.Schema)
		driver = "postgres"
	case helper.Oracle:
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

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) logQuery(query string) {
	if d.profile == "dev" {
		log.Printf("Query: %s\n", query)
	}
}

func (d *Database) isQueryIN(query string) bool {
	return strings.Contains(query, "IN") || strings.Contains(query, "in")
}

func (d *Database) GetList(c context.Context, dest interface{}, query string, args ...interface{}) error {
	var err error
	if d.isQueryIN(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return err
		}
	}
	query = d.db.Rebind(query)
	d.logQuery(query)

	ctx, cancel := helper.CreateCtxTimeout(c, d.timeoutSec)
	defer cancel()

	if helper.IsNilOrEmpty(args) {
		return d.db.SelectContext(ctx, dest, query)
	}

	return d.db.SelectContext(ctx, dest, query, args...)
}

func (d *Database) GetOne(c context.Context, dest interface{}, query string, args ...interface{}) error {
	var err error
	if d.isQueryIN(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return err
		}
	}
	query = d.db.Rebind(query)
	d.logQuery(query)

	ctx, cancel := helper.CreateCtxTimeout(c, d.timeoutSec)
	defer cancel()

	if helper.IsNilOrEmpty(args) {
		return d.db.GetContext(ctx, dest, query)
	}

	return d.db.GetContext(ctx, dest, query, args...)
}

func (d *Database) ExecQuery(c context.Context, query string, args ...interface{}) error {
	var err error
	if d.isQueryIN(query) {
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return err
		}
	}
	query = d.db.Rebind(query)
	d.logQuery(query)

	ctx, cancel := helper.CreateCtxTimeout(c, d.timeoutSec)
	defer cancel()

	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = d.db.ExecContext(ctx, query, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

func (d *Database) Insert(c context.Context, query string, args interface{}) error {
	d.logQuery(query)
	ctx, cancel := helper.CreateCtxTimeout(c, d.timeoutSec)
	defer cancel()

	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = d.db.NamedExecContext(ctx, query, args)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

func (d *Database) InsertedId(c context.Context, query string, args ...interface{}) (int, error) {
	var id int
	d.logQuery(query)
	ctx, cancel := helper.CreateCtxTimeout(c, d.timeoutSec)
	defer cancel()

	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return id, err
	}

	err = d.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		tx.Rollback()
		return id, err
	}
	tx.Commit()
	return id, err
}

func (d *Database) Update(c context.Context, query string, args interface{}) error {
	d.logQuery(query)
	ctx, cancel := helper.CreateCtxTimeout(c, d.timeoutSec)
	defer cancel()

	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = d.db.NamedExecContext(ctx, query, args)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}

func (d *Database) Delete(c context.Context, query string, args interface{}) error {
	d.logQuery(query)
	ctx, cancel := helper.CreateCtxTimeout(c, d.timeoutSec)
	defer cancel()

	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = d.db.NamedExecContext(ctx, query, args)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return err
}
