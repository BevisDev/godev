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
| **DBType**                 | `DBType`            | The target database type (e.g., `Postgres`, `MySQL`).                       |
| **Schema**                 | `string`            | The name of the target database/schema.                                     |
| **TimeoutSec**             | `int`               | Default timeout (in seconds) for DB operations. Defaults to **60 seconds**. |
| **Host**, **Port**         | `string`, `int`     | Server address and port.                                                    |
| **Username**, **Password** | `string`            | Login credentials.                                                          |
| **MaxOpenConns**           | `int`               | Maximum number of open connections in the pool.                             |
| **MaxIdleConns**           | `int`               | Maximum number of idle connections.                                         |
| **MaxIdleTimeSec**         | `int`               | Maximum time (seconds) a connection can remain idle.                        |
| **MaxLifeTimeSec**         | `int`               | Maximum time (seconds) a connection can be reused.                          |
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
		ShowQuery: true,
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	log.Printf("connect db success")
}

```