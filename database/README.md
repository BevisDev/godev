# Database Package (`database`)

The `database` package provides a robust and flexible abstraction layer for managing SQL database connections and
executing various database operations in Go applications. It utilizes the `sqlx` library to simplify struct-to-query
mapping and supports multiple database kinds (Postgres, MySQL, SQL Server, Oracle).

---

## 1. Configuration (`ConfigDB`)

The `ConfigDB` struct defines all parameters necessary to establish a database connection.

| Field                      | Type                | Description                                                                 |
|:---------------------------|:--------------------|:----------------------------------------------------------------------------|
| **Kind**                   | `types.KindDB`      | The target database type (e.g., `types.Postgres`, `types.MySQL`).           |
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

The `NewDB` function is the entry point for creating and configuring the database connection.

### `NewDB(*ConfigDB) (DBInterface, error)`

Initializes a new `*Database` instance, sets up the connection pool, and performs a `Ping` to verify connectivity.

```go
// Example Usage:

// 'db' is the primary object for executing queries
```

```go
package main

import (
	"github.com/BevisDev/godev/database"
	"log"
)

func main() {
	cfg := &Config{
		DBType:    database.Postgres,
		DBName:    "app_db",
		Host:      "localhost",
		Port:      5432,
		Username:  "user",
		Password:  "password",
		ShowQuery: true,
	}

	db, err := New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	log.Printf("connect db success")
}

```