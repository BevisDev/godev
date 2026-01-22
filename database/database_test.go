package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	Name  string `db:"name"`
	Email string `db:"email"`
}

// setupTestDB creates a test database instance with mock.
func setupTestDB(t *testing.T) (*Database, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err, "failed to create mock db")

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	database := &Database{
		Config: &Config{
			DBType:  SqlServer,
			Timeout: 5 * time.Second,
		},
		db: sqlxDB,
	}

	return database, mock
}

// ============================================================================
// Connection & Lifecycle Tests
// ============================================================================

func TestDatabase_Ping(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		mock.ExpectPing()

		err := db.Ping()
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("nil connection", func(t *testing.T) {
		db := &Database{db: nil}

		err := db.Ping()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ping error")
	})
}

func TestDatabase_Close(t *testing.T) {
	db, _ := setupTestDB(t)

	// Should not panic
	assert.NotPanics(t, func() {
		db.Close()
	})

	// Second close should be safe
	assert.NotPanics(t, func() {
		db.Close()
	})
}

func TestDatabase_GetDB(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	sqlxDB := db.GetDB()
	assert.NotNil(t, sqlxDB)
}

func TestDatabase_ViewQuery(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		db := &Database{
			Config: &Config{ShowQuery: true},
		}
		// ViewQuery should not panic when ShowQuery is true
		assert.NotPanics(t, func() {
			db.ViewQuery("SELECT * FROM users")
		})
	})

	t.Run("disabled", func(t *testing.T) {
		db := &Database{
			Config: &Config{ShowQuery: false},
		}
		assert.NotPanics(t, func() {
			db.ViewQuery("SELECT * FROM users")
		})
	})
}

// ============================================================================
// Utility Methods Tests
// ============================================================================

func TestDatabase_IsNoResult(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"sql.ErrNoRows", sql.ErrNoRows, true},
		{"wrapped ErrNoRows", fmt.Errorf("wrapped: %w", sql.ErrNoRows), true},
		{"other error", errors.New("some error"), false},
		{"nil error", nil, false},
	}

	db := &Database{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.IsNoResult(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabase_MustBePtr(t *testing.T) {
	db := &Database{}

	t.Run("valid pointer", func(t *testing.T) {
		var val int
		err := db.MustBePtr(&val)
		assert.NoError(t, err)
	})

	t.Run("non-pointer", func(t *testing.T) {
		err := db.MustBePtr(123)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a non-nil pointer")
	})

	t.Run("nil pointer", func(t *testing.T) {
		var val *int
		err := db.MustBePtr(val)
		assert.Error(t, err)
	})
}

func TestDatabase_FormatRow(t *testing.T) {
	tests := []struct {
		name     string
		dbType   DBType
		idx      int
		expected string
	}{
		{"SQL Server", SqlServer, 1, "@p1"},
		{"SQL Server", SqlServer, 5, "@p5"},
		{"Postgres", Postgres, 1, "$1"},
		{"Postgres", Postgres, 10, "$10"},
		{"MySQL", MySQL, 1, "?"},
		{"MySQL", MySQL, 5, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{Config: &Config{DBType: tt.dbType}}
			result := db.FormatRow(tt.idx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Execute Tests
// ============================================================================

func TestDatabase_Execute(t *testing.T) {
	t.Run("without transaction", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		query := "UPDATE users SET name = @p1 WHERE id = @p2"

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("Alice", 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := db.Execute(ctx, query, nil, "Alice", 1)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with transaction", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		query := "UPDATE users SET name = @p1 WHERE id = @p2"

		mock.ExpectBegin()
		tx, err := db.GetDB().BeginTxx(ctx, nil)
		require.NoError(t, err)

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("Alice", 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = db.Execute(ctx, query, tx, "Alice", 1)
		assert.NoError(t, err)

		mock.ExpectCommit()
		_ = tx.Commit()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDatabase_ExecuteTx(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "UPDATE users SET name = @p1 WHERE id = @p2"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("Alice", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := db.ExecuteTx(ctx, "UPDATE users SET name = ? WHERE id = ?", "Alice", 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_ExecuteSafe(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "UPDATE users SET name = @p1 WHERE id = @p2"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("Alice", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := db.ExecuteSafe(ctx, "UPDATE users SET name = ? WHERE id = ?", "Alice", 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ============================================================================
// Save Tests
// ============================================================================

func TestDatabase_Save(t *testing.T) {
	t.Run("without transaction", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		query := "INSERT INTO users (name, email) VALUES (@p1, @p2)"

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("Alice", "alice@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := db.Save(ctx, nil, "INSERT INTO users (name, email) VALUES (:name, :email)", map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with transaction", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		mock.ExpectBegin()
		tx, err := db.GetDB().BeginTxx(ctx, nil)
		require.NoError(t, err)

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (@p1, @p2)")).
			WithArgs("Alice", "alice@example.com").
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = db.Save(ctx, tx, "INSERT INTO users (name, email) VALUES (:name, :email)", map[string]interface{}{
			"name":  "Alice",
			"email": "alice@example.com",
		})
		assert.NoError(t, err)

		mock.ExpectCommit()
		_ = tx.Commit()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDatabase_SaveTx(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (@p1, @p2)")).
		WithArgs("Alice", "alice@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.SaveTx(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_SaveSafe(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (@p1, @p2)")).
		WithArgs("Alice", "alice@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.SaveSafe(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_InsertOrUpdate(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (@p1, @p2)")).
		WithArgs("Alice", "alice@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := db.InsertOrUpdate(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
	})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ============================================================================
// Query Tests
// ============================================================================

func TestDatabase_GetAny(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		var user User

		mock.ExpectQuery(regexp.QuoteMeta("SELECT name, email FROM users WHERE id = @p1")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"name", "email"}).
				AddRow("Alice", "alice@example.com"))

		err := db.GetAny(ctx, &user, "SELECT name, email FROM users WHERE id = ?", 1)
		assert.NoError(t, err)
		assert.Equal(t, "Alice", user.Name)
		assert.Equal(t, "alice@example.com", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		var user User

		mock.ExpectQuery(regexp.QuoteMeta("SELECT name, email FROM users WHERE id = @p1")).
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		err := db.GetAny(ctx, &user, "SELECT name, email FROM users WHERE id = ?", 999)
		assert.Error(t, err)
		assert.True(t, db.IsNoResult(err))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid destination", func(t *testing.T) {
		db, _ := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		var user User

		err := db.GetAny(ctx, user, "SELECT * FROM users WHERE id = ?", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a non-nil pointer")
	})
}

func TestDatabase_GetList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		var users []User

		mock.ExpectQuery(regexp.QuoteMeta("SELECT name, email FROM users")).
			WillReturnRows(sqlmock.NewRows([]string{"name", "email"}).
				AddRow("Alice", "alice@example.com").
				AddRow("Bob", "bob@example.com"))

		err := db.GetList(ctx, &users, "SELECT name, email FROM users")
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "Alice", users[0].Name)
		assert.Equal(t, "Bob", users[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		var users []User

		mock.ExpectQuery(regexp.QuoteMeta("SELECT name, email FROM users")).
			WillReturnRows(sqlmock.NewRows([]string{"name", "email"}))

		err := db.GetList(ctx, &users, "SELECT name, email FROM users")
		assert.NoError(t, err)
		assert.Empty(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ============================================================================
// Insert Tests
// ============================================================================

func TestDatabase_ExecReturningId(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		query := "INSERT INTO users (name) OUTPUT INSERTED.id VALUES (?)"

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("Alice").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

		id, err := db.ExecReturningId(ctx, query, "Alice")
		assert.NoError(t, err)
		assert.Equal(t, 123, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows error", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		query := "INSERT INTO users (name) OUTPUT INSERTED.id VALUES (?)"

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("Alice").
			WillReturnError(sql.ErrNoRows)

		id, err := db.ExecReturningId(ctx, query, "Alice")
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.True(t, db.IsNoResult(err))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDatabase_InsertBulk(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

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

		var args []interface{}
		for _, u := range users {
			args = append(args, u.Name, u.Email)
		}

		expectedQuery := buildExpectedInsertQuery(db, table, colNames, len(users))
		driverArgs := toDriverArgs(args)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
			WithArgs(driverArgs...).
			WillReturnResult(sqlmock.NewResult(1, int64(len(users))))
		mock.ExpectCommit()

		err := db.InsertBulk(ctx, table, len(users), colNames, args...)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid column count", func(t *testing.T) {
		db, _ := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()

		err := db.InsertBulk(ctx, "users", 2, []string{}, "arg1", "arg2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "column count must be greater than 0")
	})

	t.Run("argument count mismatch", func(t *testing.T) {
		db, _ := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		colNames := []string{"name", "email"}

		err := db.InsertBulk(ctx, "users", 2, colNames, "Alice", "alice@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected")
	})

	t.Run("rollback on error", func(t *testing.T) {
		db, mock := setupTestDB(t)
		defer db.Close()

		ctx := context.Background()
		table := "users"
		colNames := []string{"name", "email"}
		args := []interface{}{"Alice", "alice@example.com"}

		expectedQuery := buildExpectedInsertQuery(db, table, colNames, 1)
		driverArgs := toDriverArgs(args)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(expectedQuery)).
			WithArgs(driverArgs...).
			WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()

		err := db.InsertBulk(ctx, table, 1, colNames, args...)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insert failed")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDatabase_InsertMany(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "INSERT INTO users (name, email) VALUES (:name, :email)"
	entities := []interface{}{
		map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
		map[string]interface{}{"name": "Bob", "email": "bob@example.com"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (@p1, @p2)")).
		WithArgs("Alice", "alice@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (@p1, @p2)")).
		WithArgs("Bob", "bob@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.InsertMany(ctx, query, entities)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ============================================================================
// Update & Delete Tests
// ============================================================================

func TestDatabase_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = @p1")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := db.Delete(ctx, "DELETE FROM users WHERE id = :id", map[string]interface{}{"id": 1})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_UpdateMany(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "UPDATE users SET name = :name WHERE id = :id"
	entities := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice Updated"},
		map[string]interface{}{"id": 2, "name": "Bob Updated"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET name = @p1 WHERE id = @p2")).
		WithArgs("Alice Updated", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET name = @p1 WHERE id = @p2")).
		WithArgs("Bob Updated", 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := db.UpdateMany(ctx, query, entities)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_UpdateManySafe(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	query := "UPDATE users SET name = :name WHERE id = :id"
	entities := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice Updated"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET name = @p1 WHERE id = @p2")).
		WithArgs("Alice Updated", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := db.UpdateManySafe(ctx, query, entities)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ============================================================================
// Helper Functions
// ============================================================================

func buildExpectedInsertQuery(db *Database, table string, colNames []string, rowCount int) string {
	var placeholders []string
	colCount := len(colNames)

	for i := 0; i < rowCount; i++ {
		var paramRow []string
		for j := 1; j <= colCount; j++ {
			paramRow = append(paramRow, db.FormatRow(i*colCount+j))
		}
		placeholders = append(placeholders, "("+strings.Join(paramRow, ", ")+")")
	}

	colsJoin := strings.Join(colNames, ", ")
	placeholderStr := strings.Join(placeholders, ", ")
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, colsJoin, placeholderStr)
}

func toDriverArgs(args []interface{}) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, v := range args {
		values[i] = v
	}
	return values
}
