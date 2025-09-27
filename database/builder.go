package database

import (
	"fmt"
	"github.com/BevisDev/godev/types"
	"github.com/jmoiron/sqlx"
	"strings"
)

type QueryBuilder struct {
	db      *sqlx.DB
	kindDB  types.KindDB
	table   string
	columns []string
	where   []string
	args    []interface{}
	order   []string
	limit   int
	offset  int
	updates map[string]interface{}
	inserts map[string]interface{}
}

func (qb QueryBuilder) Select(cols ...string) QueryBuilder {
	qb.columns = cols
	return qb
}

func (qb QueryBuilder) Where(cond string, args ...interface{}) QueryBuilder {
	qb.where = append(qb.where, cond)
	qb.args = append(qb.args, args...)
	return qb
}

func (qb QueryBuilder) Order(order ...string) QueryBuilder {
	qb.order = order
	return qb
}

func (qb QueryBuilder) Build() (string, []interface{}) {
	cols := "*"
	if len(qb.columns) > 0 {
		cols = strings.Join(qb.columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, qb.table)

	if len(qb.where) > 0 {
		query += " WHERE " + strings.Join(qb.where, " AND ")
	}

	if len(qb.order) > 0 {
		ordered := strings.Join(qb.order, ", ")
		query += " ORDER BY " + ordered
	}

	if qb.limit >= 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}
	if qb.offset >= 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offset)
	}

	return query, qb.args
}
