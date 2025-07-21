package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/BevisDev/godev/types"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"regexp"
	"strings"
	"testing"
)

type User struct {
	Name  string `db:"name"`
	Email string `db:"email"`
}

func newTestDB(t *testing.T) (*Database, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	return &Database{
		DB:         sqlxDB,
		TimeoutSec: 5,
		kindDB:     types.SqlServer,
	}, mock
}

func TestDatabase_Execute(t *testing.T) {
	db, mock := newTestDB(t)

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET name = ? WHERE id = ?")).
		WithArgs("Alice", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.ExecuteTx(ctx, "UPDATE users SET name = ? WHERE id = ?", "Alice", 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_Save(t *testing.T) {
	db, mock := newTestDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (?, ?)")).
		WithArgs("Alice", "alice@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.Save(ctx, nil, "INSERT INTO users (name, email) VALUES (:name, :email)", map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertBulk_Users(t *testing.T) {
	d, mock := newTestDB(t)
	defer d.Close()

	ctx := context.Background()
	table := "users"
	colNames := []string{"name", "email"}
	users := []struct {
		Name  string
		Email string
	}{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	// Chuẩn bị args
	var args []interface{}
	for _, u := range users {
		args = append(args, u.Name, u.Email)
	}
	size := len(users)

	// Tạo expected query với placeholder MSSQL (@p1, @p2, ...)
	var placeholders []string
	for i := 0; i < size; i++ {
		var row []string
		for j := 1; j <= len(colNames); j++ {
			row = append(row, d.FormatRow(i*len(colNames)+j))
		}
		placeholders = append(placeholders, "("+strings.Join(row, ", ")+")")
	}
	expectedQuery := regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		table, strings.Join(colNames, ", "), strings.Join(placeholders, ", ")))

	// Chuẩn bị driver args (sqlmock yêu cầu driver.Value)
	var driverArgs []driver.Value
	for _, arg := range args {
		driverArgs = append(driverArgs, arg)
	}

	// Thiết lập mock kỳ vọng
	mock.ExpectBegin()
	mock.ExpectExec(expectedQuery).
		WithArgs(driverArgs...).
		WillReturnResult(sqlmock.NewResult(1, int64(size)))
	mock.ExpectCommit()

	// Thực thi test thành công
	err := d.InsertBulk(ctx, table, size, colNames, args...)
	assert.NoError(t, err)

	// Kiểm tra tất cả kỳ vọng của mock
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test trường hợp lỗi: số lượng args không khớp
	err = d.InsertBulk(ctx, table, size, colNames, args[1:]...) // Truyền thiếu args
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected")
}

func TestDatabase_ExecReturningId(t *testing.T) {
	d, mock := newTestDB(t)
	defer d.Close()

	ctx := context.Background()
	query := "INSERT INTO users (name) OUTPUT INSERTED.id VALUES (?)"
	name := "Alice"

	// Thiết lập mock
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

	// Thực thi test
	id, err := d.ExecReturningId(ctx, query, name)
	assert.NoError(t, err)
	assert.Equal(t, 123, id)

	// Kiểm tra tất cả kỳ vọng của mock
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test trường hợp lỗi: query thất bại
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name).
		WillReturnError(sql.ErrNoRows)

	id, err = d.ExecReturningId(ctx, query, name)
	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.True(t, d.IsNoResult(err), "expected sql.ErrNoRows")
}

func TestInsertBulk_RollbackOnError(t *testing.T) {
	d, mock := newTestDB(t)
	defer d.Close()

	ctx := context.Background()
	table := "users"
	colNames := []string{"name", "email"}
	args := []interface{}{"Alice", "alice@mail.com"}

	mock.ExpectBegin()

	// Mock lỗi Exec
	q := buildExpectedInsertQuery(d, table, colNames, 1)
	mock.ExpectExec(regexp.QuoteMeta(q)).
		WithArgs(toDriver(args)...).
		WillReturnError(errors.New("insert failed"))

	// Transaction rollback
	mock.ExpectRollback()

	err := d.InsertBulk(ctx, table, 1, colNames, args...)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper để build query giống InsertBatch
func buildExpectedInsertQuery(d *Database, table string, colNames []string, row int) string {
	var placeholders []string
	col := len(colNames)
	for i := 0; i < row; i++ {
		var paramRow []string
		for j := 1; j <= col; j++ {
			paramRow = append(paramRow, d.FormatRow(i*col+j))
		}
		placeholders = append(placeholders, "("+strings.Join(paramRow, ", ")+")")
	}
	colsJoin := strings.Join(colNames, ", ")
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, colsJoin, strings.Join(placeholders, ", "))
}

// Helper để convert args sang driver.Value
func toDriver(args []interface{}) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, v := range args {
		values[i] = v
	}
	return values
}
