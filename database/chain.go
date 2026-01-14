package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/BevisDev/godev/utils"
)

type Chain[T any] struct {
	*Database
	table string

	columns []string
	conds   []string
	orders  []string
	args    []interface{}
	where   []string

	top    int // for MSSQL
	limit  int
	offset int

	updates map[string]interface{}
	inserts map[string]interface{}
	values  []interface{}
}

func Builder[T any](db *Database) ChainExec[T] {
	return &Chain[T]{
		Database: db,
	}
}

func (d *Chain[T]) From(table string) ChainExec[T] {
	d.table = table
	return d
}

func (d *Chain[T]) clone() *Chain[T] {
	c := *d

	if len(c.columns) > 0 {
		c.columns = append([]string{}, d.columns...)
	}
	if len(c.conds) > 0 {
		c.conds = append([]string{}, d.conds...)
	}
	if len(c.orders) > 0 {
		c.orders = append([]string{}, d.orders...)
	}
	if len(c.args) > 0 {
		c.args = append([]interface{}{}, d.args...)
	}
	if len(c.where) > 0 {
		c.where = append([]string{}, d.where...)
	}
	if len(d.values) > 0 {
		c.values = append([]interface{}{}, d.values...)
	}

	if c.updates != nil {
		c.updates = make(map[string]interface{}, len(d.updates))
		for k, v := range d.updates {
			c.updates[k] = v
		}
	}

	if c.inserts != nil {
		c.inserts = make(map[string]interface{}, len(d.inserts))
		for k, v := range d.inserts {
			c.inserts[k] = v
		}
	}

	return &c
}

func (d *Chain[T]) Select(cols ...string) ChainExec[T] {
	c := d.clone()
	c.columns = cols
	return c
}

func (d *Chain[T]) Where(cond string, args ...interface{}) ChainExec[T] {
	c := d.clone()
	c.where = append(c.where, cond)
	c.args = append(c.args, args...)
	return c
}

func (d *Chain[T]) Top(n int) ChainExec[T] {
	c := d.clone()
	c.top = n
	return c
}

func (d *Chain[T]) Limit(n int) ChainExec[T] {
	d.limit = n
	return d
}

func (d *Chain[T]) Offset(n int) ChainExec[T] {
	d.offset = n
	return d
}

func (d *Chain[T]) OrderBy(order string) ChainExec[T] {
	c := d.clone()
	c.orders = append(c.orders, order)
	return c
}

func (d *Chain[T]) ToSql() (string, []interface{}) {
	var sb strings.Builder
	sb.WriteString("SELECT ")

	// top only support mssql
	if d.DBType == SqlServer && d.top > 0 {
		sb.WriteString(fmt.Sprintf("TOP %d ", d.top))
	}

	// build cols
	cols := "*"
	if len(d.columns) > 0 {
		cols = strings.Join(d.columns, ", ")
	}
	sb.WriteString(cols)

	// build table
	sb.WriteString(" FROM ")
	sb.WriteString(d.table)

	// build where
	if len(d.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(d.where, " AND "))
	}

	// build order
	if len(d.orders) > 0 {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(strings.Join(d.orders, ", "))
	}

	// LIMIT/OFFSET
	if d.DBType != SqlServer && d.limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", d.limit))
	}
	if d.DBType != SqlServer && d.offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", d.offset))
	}

	return sb.String(), d.args
}

// ============================================================
// =============== READ  ===================
// ============================================================

func (d *Chain[T]) getAny(c context.Context) (*T, error) {
	var obj T
	query, args := d.ToSql()

	query, newArgs, err := d.rebind(query, args...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := utils.NewCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	db := d.GetDB()
	if err := db.GetContext(ctx, &obj, query, newArgs...); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (d *Chain[T]) First(c context.Context) (*T, error) {
	result, err := d.getAny(c)
	if err != nil {
		if d.IsNoResult(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (d *Chain[T]) FindAll(c context.Context) ([]*T, error) {
	var list []*T
	query, args := d.ToSql()

	query, newArgs, err := d.rebind(query, args...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := utils.NewCtxTimeout(c, d.TimeoutSec)
	defer cancel()

	db := d.GetDB()
	if err = db.SelectContext(ctx, &list, query, newArgs...); err != nil {
		return nil, err
	}
	return list, nil
}

// ============================================================
// =============== INSERT / UPDATE / DELETE ===================
// ============================================================

func (d *Chain[T]) Insert(ctx context.Context, data any, outputs ...string) (*T, error) {
	if len(d.columns) == 0 {
		return nil, ErrMissingSelect
	}

	var (
		dest       T
		query      string
		returnCols string
		hasOutput  = len(outputs) > 0
	)

	if hasOutput {
		returnCols = strings.Join(outputs, ", ")
	}

	query = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (:%s)",
		d.table,
		strings.Join(d.columns, ", "),
		strings.Join(d.columns, ", :"),
	)

	switch d.DBType {
	case Postgres:
		if hasOutput {
			query += fmt.Sprintf(" RETURNING %s", returnCols)
		}

	case SqlServer:
		if hasOutput {
			query = fmt.Sprintf(
				"INSERT INTO %s (%s) OUTPUT INSERTED.%s VALUES (:%s)",
				d.table,
				strings.Join(d.columns, ", "),
				returnCols,
				strings.Join(d.columns, ", :"),
			)
		}

	default:
		d.ViewQuery(query)

		if !hasOutput {
			_, err := d.db.NamedExecContext(ctx, query, data)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}

		res, err := d.db.NamedExecContext(ctx, query, data)
		if err != nil {
			return nil, err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		if id > 0 {
			q := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", d.table)
			if err := d.db.GetContext(ctx, &dest, q, id); err != nil {
				return nil, err
			}
		}
		return &dest, nil
	}

	d.ViewQuery(query)

	// handle dont have outputs
	if !hasOutput {
		_, err := d.db.NamedExecContext(ctx, query, data)
		return nil, err
	}

	rows, err := d.db.NamedQueryContext(ctx, query, data)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	rt := reflect.TypeOf((*T)(nil)).Elem()
	switch rt.Kind() {
	case reflect.Struct:
		if err := rows.StructScan(&dest); err != nil {
			return nil, err
		}
	default:
		if err := rows.Scan(&dest); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &dest, nil
}

func (d *Chain[T]) Update(ctx context.Context, fields map[string]interface{}) (int64, error) {
	if len(d.columns) == 0 {
		return 0, ErrMissingSelect
	}
	if len(d.where) == 0 {
		return 0, ErrMissingWhere
	}

	var (
		setParts []string
		args     []interface{}
	)

	for k, v := range fields {
		setParts = append(setParts, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		d.table,
		strings.Join(setParts, ", "),
		strings.Join(d.where, " AND "),
	)

	// args where
	args = append(args, d.args...)

	query, args, err := d.rebind(query, args...)
	if err != nil {
		return 0, err
	}

	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Chain[T]) Delete(ctx context.Context) (int64, error) {
	if len(d.where) == 0 {
		return 0, ErrMissingWhere
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, strings.Join(d.where, " AND "))
	res, err := d.db.ExecContext(ctx, query, d.args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
