# 🌟 GoDev

**GoDev** is a powerful utility toolkit for Go developers, designed to streamline development by:

- **Simplifying common tasks** with pre-built utilities
- **Enhancing code reusability** through modular components
- **Boosting development speed and efficiency** with robust integrations

## 🚀 Getting Started

### Prerequisites

- [Go 1.23.4](https://go.dev/doc/install) or higher

### Installation

Install the GoDev toolkit with a single command:

```bash
go get github.com/BevisDev/godev@latest
```

### Quick Start

To verify your GoDev setup, create a simple program:

```go
package main

import (
	"fmt"
	"github.com/BevisDev/godev"
)

func main() {
	fmt.Println("Welcome to GoDev!")
	// Add your GoDev utility calls here
}
```

Run the program:

```bash
go run main.go
```

## 🛠️ Dependencies

GoDev integrates with the following libraries to provide a comprehensive development experience:

| Dependency     | Purpose                     | Installation Command                                  | Documentation Link                                         |
|----------------|-----------------------------|-------------------------------------------------------|------------------------------------------------------------|
| **Viper**      | Configuration management    | `go get github.com/spf13/viper`                       | [Viper Docs](https://github.com/spf13/viper)               |
| **Zap**        | High-performance logging    | `go get -u go.uber.org/zap`                           | [Zap Docs](https://github.com/uber-go/zap)                 |
| **Lumberjack** | Log rotation and management | `go get github.com/natefinch/lumberjack`              | [Lumberjack Docs](https://github.com/natefinch/lumberjack) |
| **Cron**       | Scheduled task execution    | `go get github.com/robfig/cron/v3@v3.0.0`             | [Cron Docs](https://github.com/robfig/cron)                |
| **Database**   | SQL database drivers        | See [Database Support](#database-support) section     | [SQL Drivers](https://go.dev/wiki/SQLDrivers)              |
| **Redis**      | In-memory data store        | `go get github.com/redis/go-redis/v9`                 | [Redis Docs](https://github.com/redis/go-redis)            |
| **RabbitMQ**   | Message queue integration   | `go get github.com/rabbitmq/amqp091-go`               | [RabbitMQ Docs](https://github.com/rabbitmq/amqp091-go)    |
| **Goose**      | Database migrations         | `go get github.com/pressly/goose/v3/cmd/goose@latest` | [Goose Docs](https://github.com/pressly/goose)             |
| **Gocloak**    | Keycloak integration        | `go get github.com/Nerzal/gocloak/v13`                | [Gocloak Docs](https://github.com/Nerzal/gocloak)          |
| **UUID**       | Unique ID generation        | Built-in with GoDev                                   | [UUID Docs](https://github.com/google/uuid)                |
| **Decimal**    | Decimal arithmetic          | Built-in with GoDev                                   | [Decimal Docs](https://github.com/shopspring/decimal)      |

## 📚 Core Features

### Configuration Management with Viper

GoDev uses **Viper** to load environment-specific configuration files based on the `GO_PROFILE` environment variable.
Configuration files (e.g., `dev.yml`, `prod.yml`) are typically stored in a `configs/` directory.

### Environment Configuration

Switch between environments (e.g., `dev`, `prod`) by setting the `GO_PROFILE` environment variable:

**On Windows:**

```bash
setx GO_PROFILE dev
```

**On Linux/macOS:**

```bash
export GO_PROFILE=dev
```

> **Note**: After setting the environment variable on Windows, restart your terminal to apply the changes.

**Example Configuration File (`configs/config.dev.yaml`):**

```yaml
server:
  port: 8080
  host: localhost
```

**Loading Configuration with Viper:**

```go
package main

import (
	"fmt"
	"github.com/BevisDev/godev/config"
	"os"
)

type AppConfig struct {
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`
}

func main() {
	// Get the environment from GO_PROFILE, default to "dev"
	profile := os.Getenv("GO_PROFILE")
	if profile == "" {
		profile = "dev"
	}

	// mapping your struct
	var appConfig AppConfig

	cf := &config.Config{
		Path:       "./configs",
		ConfigType: "yml",
		Dest:       &appConfig,
		Profile:    profile,
	}
	err := config.NewConfig(cf)
	if err != nil {
		err = fmt.Errorf("error load config %v", err)
		return
	}

	fmt.Printf("Server Port: %d\n", appConfig.Server.Port)
	fmt.Printf("Server Host: %s\n", appConfig.Server.Host)
}
```

> **Tip**: Viper supports multiple formats (YAML, JSON, TOML, etc.). Adjust `SetConfigType` accordingly if using a
> different format.

### Cron Jobs

The `cron` package enables scheduling tasks using a flexible 6-field cron expression format (including seconds).

**Example:**

```go
package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
)

func main() {
	c := cron.New(cron.WithSeconds())
	// second minute hour day month weekday
	c.AddFunc("0 * * * * *", func() {
		fmt.Println("Running every minute at the 0th second!")
	})
	c.Start()
	select {} // Keep the program running
}
```

**Cron Expression Format:**

| Field        | Mandatory | Allowed Values  | Special Characters |
|--------------|-----------|-----------------|--------------------|
| Seconds      | Yes       | 0-59            | `* / , -`          |
| Minutes      | Yes       | 0-59            | `* / , -`          |
| Hours        | Yes       | 0-23            | `* / , -`          |
| Day of Month | Yes       | 1-31            | `* / , - ?`        |
| Month        | Yes       | 1-12 or JAN-DEC | `* / , -`          |
| Day of Week  | Yes       | 0-6 or SUN-SAT  | `* / , - ?`        |

### Database Support

GoDev supports multiple SQL databases. Install the appropriate driver:

- **SQL Server**: `go get github.com/denisenkom/go-mssqldb`
- **PostgreSQL**: `go get github.com/lib/pq`
- **Oracle**: `go get github.com/godror/godror@latest`

For simplified struct mapping, use `sqlx`:

```bash
go get github.com/jmoiron/sqlx
```

### Database Migrations with Goose

GoDev uses **Goose** for database migrations to manage schema changes.

**Installation:**

```bash
go get github.com/pressly/goose/v3/cmd/goose@latest
```

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

### RabbitMQ Integration

GoDev supports RabbitMQ for message queue operations. Key methods:

- `ack(multiple bool)`: Acknowledge one (`false`) or all (`true`) messages.
- `nack(multiple, requeue bool)`: Negative acknowledgment with requeue option.

### Install Make

To use build automation scripts, install `make`:

First, install Chocolatey (run in PowerShell with Administrator privileges):

```bash
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

Then, install `make`:

```bash
choco install make
```

**On Linux (using apt):**

```bash
sudo apt update
sudo apt install make
```

**On macOS (using Homebrew):**

```bash
brew install make
```

## 🧑‍💻 Contributing

Contributions are welcome! To contribute:

1. Fork the [GoDev repository](https://github.com/BevisDev/godev).
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -m "Add your feature"`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Open a pull request.

Please include tests and documentation for new features.

## 📜 License

This project is licensed under the [MIT License](LICENSE).