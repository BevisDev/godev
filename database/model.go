package database

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/BevisDev/godev/utils"
)

// TableNamer allows a model to define its table name.
type TableNamer interface {
	TableName() string
}

type modelChain[T any] struct {
	*DB
	table    string
	tableErr error
	where    []string
	args     []interface{}
}

// Model creates a new model chain based on TableName() from type T.
func Model[T any](db *DB) ModelExec[T] {
	table, err := tableNameFor[T]()
	return &modelChain[T]{
		DB:       db,
		table:    table,
		tableErr: err,
	}
}

func (m *modelChain[T]) clone() *modelChain[T] {
	c := *m
	if len(m.where) > 0 {
		c.where = append([]string{}, m.where...)
	}
	if len(m.args) > 0 {
		c.args = append([]interface{}{}, m.args...)
	}
	return &c
}

func (m *modelChain[T]) ensureTable() error {
	if m.tableErr != nil {
		return m.tableErr
	}
	if strings.TrimSpace(m.table) == "" {
		return ErrMissingTable
	}
	return nil
}

func (m *modelChain[T]) Where(cond string, args ...interface{}) ModelExec[T] {
	c := m.clone()
	c.where = append(c.where, cond)
	c.args = append(c.args, args...)
	return c
}

func (m *modelChain[T]) First(ctx context.Context) (*T, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	query, args := m.buildSelect(1)
	query, args, err := m.rebind(query, args...)
	if err != nil {
		return nil, err
	}

	var obj T
	cctx, cancel := utils.NewCtxTimeout(ctx, m.Timeout)
	defer cancel()

	if err := m.db.GetContext(cctx, &obj, query, args...); err != nil {
		if m.IsNoResult(err) {
			return nil, nil
		}
		return nil, err
	}
	return &obj, nil
}

func (m *modelChain[T]) Find(ctx context.Context) ([]*T, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}

	query, args := m.buildSelect(0)
	query, args, err := m.rebind(query, args...)
	if err != nil {
		return nil, err
	}

	var list []*T
	cctx, cancel := utils.NewCtxTimeout(ctx, m.Timeout)
	defer cancel()

	if err := m.db.SelectContext(cctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *modelChain[T]) Create(ctx context.Context, data any) (*T, error) {
	if err := m.ensureTable(); err != nil {
		return nil, err
	}
	cols, vals, err := extractColumnsAndValues(data)
	if err != nil {
		return nil, err
	}

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		m.table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	// Databases with RETURNING/OUTPUT support
	switch m.DBType {
	case Postgres:
		query += " RETURNING *"
	case SqlServer:
		query = fmt.Sprintf(
			"INSERT INTO %s (%s) OUTPUT INSERTED.* VALUES (%s)",
			m.table,
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
		)
	}

	query, vals, err = m.rebind(query, vals...)
	if err != nil {
		return nil, err
	}

	cctx, cancel := utils.NewCtxTimeout(ctx, m.Timeout)
	defer cancel()

	// If the DB supports RETURNING/OUTPUT, fetch the inserted row.
	if m.DBType == Postgres || m.DBType == SqlServer {
		var dest T
		row := m.db.QueryRowxContext(cctx, query, vals...)
		if err := row.StructScan(&dest); err != nil {
			return nil, err
		}
		return &dest, nil
	}

	// Default: execute and return nil (or fetch by id when available).
	res, err := m.db.ExecContext(cctx, query, vals...)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil || id <= 0 {
		return nil, nil
	}

	// Best-effort fetch by "id"
	var dest T
	q := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", m.table)
	q, args, err := m.rebind(q, id)
	if err != nil {
		return nil, err
	}
	if err := m.db.GetContext(cctx, &dest, q, args...); err != nil {
		return nil, err
	}
	return &dest, nil
}

func (m *modelChain[T]) Updates(ctx context.Context, data any) (int64, error) {
	if err := m.ensureTable(); err != nil {
		return 0, err
	}
	if len(m.where) == 0 {
		return 0, ErrMissingWhere
	}
	cols, vals, err := extractColumnsAndValues(data)
	if err != nil {
		return 0, err
	}

	setParts := make([]string, len(cols))
	for i, col := range cols {
		setParts[i] = fmt.Sprintf("%s = ?", col)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		m.table,
		strings.Join(setParts, ", "),
		strings.Join(m.where, " AND "),
	)

	vals = append(vals, m.args...)
	query, vals, err = m.rebind(query, vals...)
	if err != nil {
		return 0, err
	}

	cctx, cancel := utils.NewCtxTimeout(ctx, m.Timeout)
	defer cancel()

	res, err := m.db.ExecContext(cctx, query, vals...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (m *modelChain[T]) Delete(ctx context.Context) (int64, error) {
	if err := m.ensureTable(); err != nil {
		return 0, err
	}
	if len(m.where) == 0 {
		return 0, ErrMissingWhere
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		m.table,
		strings.Join(m.where, " AND "),
	)

	query, args, err := m.rebind(query, m.args...)
	if err != nil {
		return 0, err
	}

	cctx, cancel := utils.NewCtxTimeout(ctx, m.Timeout)
	defer cancel()

	res, err := m.db.ExecContext(cctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (m *modelChain[T]) Count(ctx context.Context) (int64, error) {
	if err := m.ensureTable(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf("SELECT COUNT(1) FROM %s", m.table)
	if len(m.where) > 0 {
		query += " WHERE " + strings.Join(m.where, " AND ")
	}

	query, args, err := m.rebind(query, m.args...)
	if err != nil {
		return 0, err
	}

	cctx, cancel := utils.NewCtxTimeout(ctx, m.Timeout)
	defer cancel()

	var count int64
	if err := m.db.GetContext(cctx, &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *modelChain[T]) buildSelect(limit int) (string, []interface{}) {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	if m.DBType == SqlServer && limit > 0 {
		sb.WriteString(fmt.Sprintf("TOP %d ", limit))
	}
	sb.WriteString("* FROM ")
	sb.WriteString(m.table)

	if len(m.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(m.where, " AND "))
	}

	if m.DBType != SqlServer && limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	}
	return sb.String(), m.args
}

func tableNameFor[T any]() (string, error) {
	var zero T
	candidates := []any{zero}

	v := reflect.ValueOf(zero)
	if v.IsValid() && v.Kind() == reflect.Ptr && v.IsNil() {
		candidates = append(candidates, reflect.New(v.Type().Elem()).Interface())
	} else if v.IsValid() && v.Kind() != reflect.Ptr {
		candidates = append(candidates, &zero)
	}

	for _, c := range candidates {
		if tn, ok := c.(TableNamer); ok {
			name := strings.TrimSpace(tn.TableName())
			if name == "" {
				return "", ErrMissingTable
			}
			return name, nil
		}
	}
	return "", ErrMissingTable
}

func extractColumnsAndValues(data any) ([]string, []interface{}, error) {
	if data == nil {
		return nil, nil, ErrMissingData
	}

	v := reflect.ValueOf(data)
	for v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil, ErrMissingData
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return nil, nil, fmt.Errorf("[database] map key must be string")
		}
		keys := v.MapKeys()
		cols := make([]string, 0, len(keys))
		for _, k := range keys {
			cols = append(cols, k.String())
		}
		sort.Strings(cols)

		vals := make([]interface{}, 0, len(cols))
		for _, k := range cols {
			vals = append(vals, v.MapIndex(reflect.ValueOf(k)).Interface())
		}
		return cols, vals, nil

	case reflect.Struct:
		t := v.Type()
		cols := make([]string, 0, t.NumField())
		vals := make([]interface{}, 0, t.NumField())

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}

			tag := strings.TrimSpace(f.Tag.Get("db"))
			if tag == "-" {
				continue
			}
			if tag != "" {
				tag = strings.Split(tag, ",")[0]
			}
			if tag == "" {
				tag = f.Name
			}

			cols = append(cols, tag)
			vals = append(vals, v.Field(i).Interface())
		}

		if len(cols) == 0 {
			return nil, nil, fmt.Errorf("[database] no exported fields to map")
		}
		return cols, vals, nil
	default:
		return nil, nil, fmt.Errorf("[database] unsupported data type: %s", v.Kind())
	}
}
