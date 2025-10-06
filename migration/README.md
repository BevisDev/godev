# Migration Package

The `migration` package provides a standardized interface and utilities to manage database schema migrations in Go
applications.

---

## Interface

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/BevisDev/godev/migration"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=testdb sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	migConfig := &migration.Config{
		Dir:     "./migrations",
		DBType:  migration.Postgres,
		DB:      db,
		Timeout: 30,
	}

	m, err := migration.New(migConfig)
	if err != nil {
		log.Fatal("Failed to initialize migration:", err)
	}

	ctx := context.Background()

	// Apply all migrations
	if err := m.Up(ctx, 0); err != nil {
		log.Fatal("Migration up failed:", err)
	}

	// Check status
	if err := m.Status(); err != nil {
		log.Println("Status error:", err)
	}

	// Rollback last migration
	if err := m.Down(ctx, 0); err != nil {
		log.Fatal("Migration down failed:", err)
	}

	fmt.Println("Migration operations completed successfully!")
}

```