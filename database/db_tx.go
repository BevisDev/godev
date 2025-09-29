package database

import "github.com/jmoiron/sqlx"

type DBTx struct {
	db *sqlx.DB
}
