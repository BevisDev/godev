# Database Package (`database`)

The `database` package provides a robust and flexible abstraction layer for managing SQL database connections and
executing various database operations in Go applications. It utilizes the [sqlx](https://github.com/jmoiron/sqlx)
library to simplify struct-to-query
mapping and supports multiple database kinds (Postgres, MySQL, SQL Server, Oracle).

---

## Database Support

GoDev supports multiple SQL databases. Install the appropriate driver:

- **SQL Server**: `go get github.com/denisenkom/go-mssqldb`
- **PostgreSQL**: `go get github.com/lib/pq`
- **Oracle**: `go get github.com/godror/godror@latest`
- **Other**: [SQL Drivers](https://go.dev/wiki/SQLDrivers)

## 1. Configuration (`Config`)

The `Config` struct defines all parameters necessary to establish a database connection.

| Field                      | Type                | Description                                                                 |
|:---------------------------|:--------------------|:----------------------------------------------------------------------------|
| **DBType**                 | `DBType`            | The target database type (e.g., `Postgres`, `MySQL`, `SqlServer`, `Oracle`). |
| **DBName**                 | `string`            | The name of the target database/schema.                                     |
| **Timeout**                | `time.Duration`     | Default timeout for DB operations. Defaults to **60 seconds**.              |
| **Host**, **Port**         | `string`, `int`     | Server address and port.                                                    |
| **Username**, **Password** | `string`            | Login credentials.                                                          |
| **MaxOpenConns**           | `int`               | Maximum number of open connections in the pool. Defaults to **50**.         |
| **MaxIdleConns**           | `int`               | Maximum number of idle connections. Defaults to **50**.                    |
| **MaxIdleTime**            | `time.Duration`     | Maximum time a connection can remain idle. Defaults to **5 seconds**.       |
| **MaxLifeTime**            | `time.Duration`     | Maximum time a connection can be reused. Defaults to **3600 seconds**.      |
| **ShowQuery**              | `bool`              | Enables logging of executed SQL queries.                                    |
| **Params**                 | `map[string]string` | Optional additional parameters for the connection string.                   |

---

## 2. Initialization

The `New` function is the entry point for creating and configuring the database connection.

### `New(*Config) (*Database, error)`

Initializes a new `*Database` instance, sets up the connection pool, and performs a `Ping` to verify connectivity.

```go
package main

import (
	"github.com/BevisDev/godev/database"
	"log"
	
	// import driver sql
	_ "github.com/lib/pq"
)

func main() {
	cfg := &database.Config{
		DBType:    database.Postgres,
		DBName:    "app_db",
		Host:      "localhost",
		Port:      5432,
		Username:  "user",
		Password:  "password",
		Timeout:   5 * time.Second,
		ShowQuery: true,
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer db.Close()

	log.Printf("Connected to database successfully")
}

```