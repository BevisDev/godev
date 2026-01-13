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

| Dependency  | Purpose                  | Installation Command                      | Documentation Link                                    |
|-------------|--------------------------|-------------------------------------------|-------------------------------------------------------|
| **Cron**    | Scheduled task execution | `go get github.com/robfig/cron/v3@v3.0.0` | [Cron Docs](https://github.com/robfig/cron)           |
| **UUID**    | Unique ID generation     | Built-in with GoDev                       | [UUID Docs](https://github.com/google/uuid)           |
| **Decimal** | Decimal arithmetic       | Built-in with GoDev                       | [Decimal Docs](https://github.com/shopspring/decimal) |

## ✨ Features
- Scheduler

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