package database

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

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

	err := db.Execute(ctx, "UPDATE users SET name = ? WHERE id = ?", "Alice", 1)
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

	err := db.Save(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
	})
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
