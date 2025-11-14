# Migration Package

The `migration` package provides a standardized interface and utilities to manage database schema migrations in Go
applications.

---

### Installation

```bash
go get github.com/pressly/goose/v3/cmd/goose@latest
```

**See more**: [Goose Docs](https://github.com/pressly/goose)

**Example Migration File (`migrations/00001_create_users_table.sql`):**

```sql
-- +goose Up
CREATE TABLE users
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
```

**Running Migrations:**

1. Create a `migrations/` directory in your project.
2. Add SQL migration files with `Up` and `Down` directives as shown above.
3. Run migrations using the `goose` CLI or using `migration package` in code

```bash
goose -dir migrations postgres "user=postgres password=secret dbname=mydb sslmode=disable" up
```

**Commands:**

- `goose up`: Apply all available migrations.
- `goose down`: Roll back the latest migration.
- `goose status`: Check the status of migrations.

> **Tip**: Ensure the database connection string matches your configuration in `config.<env>.yaml`.

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