package database

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ModelUser struct {
	Name  string `db:"name"`
	Email string `db:"email"`
}

func (ModelUser) TableName() string { return "users" }

func TestModel_First(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT TOP 1 * FROM users WHERE id = @p1"),
	).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"name", "email"}).
			AddRow("Alice", "alice@example.com"))

	user, err := Model[ModelUser](db).
		Where("id = ?", 1).
		First(ctx)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModel_Find(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT * FROM users WHERE age > @p1"),
	).
		WithArgs(18).
		WillReturnRows(sqlmock.NewRows([]string{"name", "email"}).
			AddRow("Alice", "alice@example.com").
			AddRow("Bob", "bob@example.com"))

	users, err := Model[ModelUser](db).
		Where("age > ?", 18).
		Find(ctx)

	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModel_Count(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectQuery(
		regexp.QuoteMeta("SELECT COUNT(1) FROM users WHERE age > @p1"),
	).
		WithArgs(18).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	count, err := Model[ModelUser](db).
		Where("age > ?", 18).
		Count(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModel_Create_SqlServer(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectQuery(
		regexp.QuoteMeta("INSERT INTO users (name, email) OUTPUT INSERTED.* VALUES (@p1, @p2)"),
	).
		WithArgs("Alice", "alice@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"name", "email"}).
			AddRow("Alice", "alice@example.com"))

	user, err := Model[ModelUser](db).
		Create(ctx, ModelUser{Name: "Alice", Email: "alice@example.com"})

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "Alice", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModel_Updates(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectExec(
		regexp.QuoteMeta("UPDATE users SET name = @p1 WHERE id = @p2"),
	).
		WithArgs("Alice Updated", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows, err := Model[ModelUser](db).
		Where("id = ?", 1).
		Updates(ctx, map[string]interface{}{"name": "Alice Updated"})

	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModel_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectExec(
		regexp.QuoteMeta("DELETE FROM users WHERE id = @p1"),
	).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows, err := Model[ModelUser](db).
		Where("id = ?", 1).
		Delete(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestModel_Updates_MissingWhere(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	_, err := Model[ModelUser](db).Updates(ctx, map[string]interface{}{"name": "Alice"})
	assert.Error(t, err)
	assert.Equal(t, ErrMissingWhere, err)
}

func TestModel_Delete_MissingWhere(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	_, err := Model[ModelUser](db).Delete(ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrMissingWhere, err)
}
