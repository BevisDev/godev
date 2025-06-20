package database

import (
	"context"
	"database/sql/driver"
	"fmt"
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

func TestInsertMany_Users(t *testing.T) {
	db, mock := newTestDB(t)
	ctx := context.Background()

	table := "users"
	colNames := []string{"name", "email"}
	users := []User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	var args []interface{}
	for _, u := range users {
		args = append(args, u.Name, u.Email)
	}

	size := len(users)

	// Build expected query
	var placeholders []string
	for i := 0; i < size; i++ {
		var row []string
		for j := 1; j <= len(colNames); j++ {
			row = append(row, fmt.Sprintf("@p%d", i*len(colNames)+j))
		}
		placeholders = append(placeholders, "("+strings.Join(row, ", ")+")")
	}

	mockPlaceholder := make([]string, size)
	for i := 0; i < size; i++ {
		mockPlaceholder[i] = "(?, ?)"
	}

	var driverArgs []driver.Value
	for _, arg := range args {
		driverArgs = append(driverArgs, arg)
	}
	expectedQuery := "INSERT INTO users (name, email) VALUES (@p1, @p2), (@p3, @p4)"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
		WithArgs(driverArgs...).
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := db.InsertMany(ctx, table, size, colNames, args...)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_InsertedId(t *testing.T) {
	db, mock := newTestDB(t)

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("Alice").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))
	mock.ExpectCommit()

	id, err := db.InsertedId(ctx, "INSERT INTO users (name) OUTPUT INSERTED.ID VALUES (?)", "Alice")
	assert.NoError(t, err)
	assert.Equal(t, 123, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}
