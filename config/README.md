# Config Package

The `config` package provides utilities for loading application configuration
from files and environment variables.  
It is built on top of Viper and adds support for:

- strongly-typed struct binding
- environment variable overrides
- `$VAR` placeholder expansion
- nested struct + slice handling.

---

## Description

GoDev uses **Viper** to load configuration files based on an active *profile*
(often controlled via the `GO_PROFILE` environment variable)

Typical config files live inside a `configs/` directory:

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

## Features

- Load config from file (yaml, json, toml, …)
- Environment variable overrides (APP_PORT overrides app.port)
- Expand placeholders like $DB_DSN
- Recursive replacement across nested maps/slices
- Type conversion support:
    - `string`, `int`, `float`, `bool`
    - slices of those types (comma-separated)
- Struct-based configuration using `config` struct tags.
- Safe pointer checks to prevent runtime panics.
- MustLoad() helper that panics on startup failure

---

**Example Configuration File (`configs/config.dev.yaml`):**

```yaml
server:
  port: 8080
  host: localhost
```

## Structure

### `Config`

Main struct for loading configuration:

| Field        | Description                                                                    |
|--------------|--------------------------------------------------------------------------------|
| `Path`       | Directory where the config file is located. Example: `"./configs"`.            |
| `Extension`  | File type: `"yaml"`, `"json"`, `"toml"`, etc.                                  |
| `Target`     | Generic struct type to receive parsed values.                                  |
| `AutoEnv`    | Enable automatic environment variable binding via Viper.                       |
| `ReplaceEnv` | Replace `$VAR` placeholders in config values with environment variable values. |
| `Profile`    | Config file name (without extension), e.g., `"dev"` or `"prod"`.               |

---

## Functions

### `MustLoad[T](cfg *Config) Response[T]`

Loads configuration and panics if loading fails

**Example:**

```go
package main

import (
	"log"
	"os"

	"github.com/BevisDev/godev/config"
)

type AppConfig struct {
	Port int    `mapstructure:"port"`
	Name string `mapstructure:"name"`
}

func main() {
	profile := os.Getenv("GO_PROFILE") // e.g., "dev" or "prod"

	result := config.MustLoad[*AppConfig](&config.Config{
		Path:      "./configs",
		Extension: "yaml",
		Profile:   profile,
	})

	log.Printf("Loaded config: %+v", result.Data)
}

```