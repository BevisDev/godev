# Config Package

The `config` package provides utilities for loading application configuration from files and environment variables.  
It is designed to work with Viper and supports struct binding, environment variable expansion, and flexible type
handling.

---

## Description

GoDev uses [Viper](https://github.com/spf13/viper) to load environment-specific configuration files based on
the `GO_PROFILE` environment variable.
Configuration files (e.g., `dev.yml`, `prod.yml`) are typically stored in a `configs/` directory.

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

- Load configuration from file (`yaml`, `json`, `toml`, etc.).
- Automatically bind environment variables to config keys.
- Expand environment variable placeholders in config values (e.g., `$APP_NAME`).
- Recursive parsing and replacement for nested structs, slices, and maps.
- Type conversion support for:
    - `string`, `int`, `float`, `bool`
    - Slices of the above types (comma-separated)
- Struct-based configuration using `config` struct tags.
- Safe pointer checks to prevent runtime panics.

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
| `ConfigType` | Type of the config file: `"yaml"`, `"json"`, `"toml"`, etc.                    |
| `Dest`       | Pointer to a struct to receive the parsed configuration. Must be a pointer.    |
| `AutoEnv`    | Enable automatic environment variable binding via Viper.                       |
| `ReplaceEnv` | Replace `$VAR` placeholders in config values with environment variable values. |
| `Profile`    | Name of the config file without extension, e.g., `"dev"` or `"prod"`.          |

---

## Functions

### `NewConfig(cf *Config) error`

Loads configuration from file and optionally merges environment variables.

**Example:**

```go
package main

import (
	"log"
	"os"
	"github.com/BevisDev/godev/config"
)

type AppConfig struct {
	Port int    `config:"app.port"`
	Name string `config:"app.name"`
}

func main() {
	profile := os.Getenv("GO_PROFILE") // e.g., "dev" or "prod"

	var cfg AppConfig
	err := config.NewConfig(&config.Config{
		Path:       "./configs",
		ConfigType: "yaml",
		Dest:       &cfg,
		Profile:    profile,
		AutoEnv:    true,
		ReplaceEnv: true,
	})
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Loaded config: %+v", cfg)
}
```