package database

import (
	"context"
	"fmt"
	"github.com/BevisDev/godev/utils"
	"strings"
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

func Query[T any](db *Database) ChainExec[T] {
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
	c.columns = append([]string{}, d.columns...)
	c.conds = append([]string{}, d.conds...)
	c.orders = append([]string{}, d.orders...)
	c.args = append([]interface{}{}, d.args...)
	c.where = append([]string{}, d.where...)

	c.updates = make(map[string]interface{}, len(d.updates))
	for k, v := range d.updates {
		c.updates[k] = v
	}

	c.inserts = make(map[string]interface{}, len(d.inserts))
	for k, v := range d.inserts {
		c.inserts[k] = v
	}

	c.values = append([]interface{}{}, d.values...)
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

func (d *Chain[T]) build() (string, []interface{}) {
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
		sb.WriteString(strings.Join(d.orders, " AND "))
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
	query, args := d.build()

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
	return d.getAny(c)
}

func (d *Chain[T]) FirstOrNil(c context.Context) (*T, error) {
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
	query, args := d.build()

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

func (d *Chain[T]) Insert(ctx context.Context, data any) (*T, error) {
	if len(d.columns) == 0 {
		return nil, fmt.Errorf("insert: missing columns — please use Select(...) before Insert")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (:%s)",
		d.table,
		strings.Join(d.columns, ", "),
		strings.Join(d.columns, ", :"),
	)

	var dest T
	switch d.DBType {
	case Postgres:
		query += " RETURNING *"

		rows, err := d.db.NamedQueryContext(ctx, query, data)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if rows.Next() {
			if err := rows.StructScan(&dest); err != nil {
				return nil, err
			}
		}
		return &dest, nil

	case SqlServer:
		query += " OUTPUT INSERTED.*"
		d.ViewQuery(query)

		rows, err := d.db.NamedQueryContext(ctx, query, data)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		if rows.Next() {
			if err := rows.StructScan(&dest); err != nil {
				return nil, err
			}
		}
		return &dest, nil

	default:
		res, err := d.db.NamedExecContext(ctx, query, data)
		if err != nil {
			return nil, err
		}
		id, _ := res.LastInsertId()

		// Nếu có ID, query lại
		if id > 0 {
			q := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", d.table)
			if err := d.db.GetContext(ctx, &dest, q, id); err != nil {
				return nil, err
			}
		}
		return &dest, nil
	}
}

func (d *Chain[T]) Update(ctx context.Context, fields map[string]interface{}) (int64, error) {
	if len(fields) == 0 {
		return 0, fmt.Errorf("no fields to update")
	}

	setParts := []string{}
	args := []interface{}{}

	for k, v := range fields {
		setParts = append(setParts, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", d.table, strings.Join(setParts, ", "))

	if len(d.where) > 0 {
		query += " WHERE " + strings.Join(d.where, " AND ")
		args = append(args, d.args...)
	}

	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Chain[T]) Delete(ctx context.Context) (int64, error) {
	if len(d.where) == 0 {
		return 0, fmt.Errorf("delete without WHERE is not allowed")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, strings.Join(d.where, " AND "))
	res, err := d.db.ExecContext(ctx, query, d.args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
