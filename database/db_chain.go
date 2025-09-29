package database

import (
	"context"
	"fmt"
	"github.com/BevisDev/godev/utils"
	"strings"
)

type DBChain[T any] struct {
	base  *Database
	table string

	columns []string
	conds   []string
	orders  []string
	args    []interface{}
	where   []string

	top int

	limit  int
	offset int

	updates map[string]interface{}
	inserts map[string]interface{}
	values  []interface{}
}

func Query[T any](db *Database) DBChainExec[T] {
	return &DBChain[T]{
		base: db,
	}
}

func (d *DBChain[T]) From(table string) DBChainExec[T] {
	d.table = table
	return d
}

func (d *DBChain[T]) clone() *DBChain[T] {
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

func (d *DBChain[T]) Select(cols ...string) DBChainExec[T] {
	c := d.clone()
	c.columns = cols
	return c
}

func (d *DBChain[T]) Top(n int) DBChainExec[T] {
	c := d.clone()
	c.top = n
	return c
}

func (d *DBChain[T]) Where(cond string, args ...interface{}) DBChainExec[T] {
	c := d.clone()
	c.where = append(c.where, cond)
	c.args = append(c.args, args...)
	return c
}

func (d *DBChain[T]) OrderBy(order string) DBChainExec[T] {
	c := d.clone()
	c.orders = append(c.orders, order)
	return c
}

func (d *DBChain[T]) build() (string, []interface{}) {
	cols := "*"
	if len(d.columns) > 0 {
		cols = strings.Join(d.columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, d.table)

	if len(d.where) > 0 {
		query += " WHERE " + strings.Join(d.where, " AND ")
	}
	if len(d.orders) > 0 {
		query += " ORDER BY " + strings.Join(d.orders, ", ")
	}

	if d.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", d.limit)
	}
	if d.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", d.offset)
	}
	return query, d.args
}

func (d *DBChain[T]) getAny(c context.Context) (*T, error) {
	var obj T
	query, args := d.build()
	base := d.base

	query, newArgs, err := base.rebind(query, args...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := utils.NewCtxTimeout(c, base.timeout)
	defer cancel()

	db := base.GetDB()
	if err := db.GetContext(ctx, &obj, query, newArgs...); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (d *DBChain[T]) First(c context.Context) (*T, error) {
	return d.getAny(c)
}

func (d *DBChain[T]) FirstOrNil(c context.Context) (*T, error) {
	result, err := d.getAny(c)
	if err != nil {
		if d.base.IsNoResult(err) {
			return nil, nil
		}
		return nil, err
	}

	return result, nil
}

func (d *DBChain[T]) FindAll(c context.Context) ([]*T, error) {
	var list []*T
	base := d.base
	query, args := d.build()

	query, newArgs, err := base.rebind(query, args...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := utils.NewCtxTimeout(c, base.timeout)
	defer cancel()

	db := base.GetDB()
	if err = db.SelectContext(ctx, &list, query, newArgs...); err != nil {
		return nil, err
	}
	return list, nil
}

func (d *DBChain[T]) Insert(columns []string, values []interface{}) DBChainExec[T] {
	d.columns = columns
	d.values = values
	return d
}
