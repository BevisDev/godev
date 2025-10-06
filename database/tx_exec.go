package database

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type DBTxExec interface {
	RunTx(c context.Context, level sql.IsolationLevel,
		fn func(ctx context.Context, tx *sqlx.Tx) error,
	) error
}
