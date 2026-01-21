# Logx Package

The `logx` package provides a structured logging solution for Go applications using
the [Zap](https://github.com/uber-go/zap) logging library.

It supports file-based logging with rotation, console logging, structured JSON logging, and flexible caller skip
configuration.

It also includes convenient methods for logging HTTP requests and responses.

---

## Features

- Structured logging with Zap (JSON or console format).
- File-based logging with rotation via [Lumberjack](https://github.com/natefinch/lumberjack):
    - `MaxSize` (MB per file)
    - `MaxBackups` (number of rotated files)
    - `MaxAge` (days)
    - `Compress` (gzip old logs)
- Supports profiles (`dev`, `prod`) to configure encoder and output.
- Caller skip configuration for request/response logs (internal and external).
- Log HTTP requests/responses with structured fields:
    - `State`, `URL`, `Method`, `Query`, `Header`, `Body`, `Status`, `Duration`.
- Supports custom logging levels: `Info`, `Warn`, `Error`, `Panic`, `Fatal`.
- Automatic JSON serialization for complex types.
- Handles `sql.Null*`, `decimal.Decimal`, `time.Time`, and raw bytes.

---

## Structure

### `Config`

Configuration for creating a logger:

| Field          | Description                                          |
|----------------|------------------------------------------------------|
| `IsProduction` | Runtime profile `"prod"`.                            |
| `IsLocal`      | Runtime profile `"dev"`, it logs output to console   |
| `MaxSize`      | Maximum size of a log file in MB before rotation.    |
| `MaxBackups`   | Maximum number of old log files to keep.             |
| `MaxAge`       | Maximum number of days to retain logs.               |
| `Compress`     | Whether to gzip old log files.                       |
| `IsRotate`     | Whether to rotate log files daily or by module.      |
| `DirName`      | Directory path to store log files.                   |
| `Filename`     | Base log filename, e.g., `"app.log"`.                |
| `CallerConfig` | Caller skip configuration for request/response logs. |

### `AppLogger`

Main struct for logging, wrapping a `*zap.Logger`.

Key methods:

| Method                                 | Description                    |
|----------------------------------------|--------------------------------|
| `Info(state, msg, args...)`            | Log an info message.           |
| `Warn(state, msg, args...)`            | Log a warning message.         |
| `Error(state, msg, args...)`           | Log an error message.          |
| `Panic(state, msg, args...)`           | Log a panic message.           |
| `Fatal(state, msg, args...)`           | Log a fatal message.           |
| `LogRequest(req *RequestLogger)`       | Log an internal request.       |
| `LogResponse(resp *ResponseLogger)`    | Log an internal response.      |
| `LogExtRequest(req *RequestLogger)`    | Log an external request.       |
| `LogExtResponse(resp *ResponseLogger)` | Log an external response.      |
| `Sync()`                               | Flush buffered logs to output. |

### `RequestLogger` / `ResponseLogger`

Structs used to log HTTP requests and responses:

| Struct           | Key Fields                                                         |
|------------------|--------------------------------------------------------------------|
| `RequestLogger`  | `State`, `URL`, `Method`, `Query`, `Header`, `Body`, `RequestTime` |
| `ResponseLogger` | `State`, `Status`, `Header`, `Body`, `DurationSec`                 |

---

```go
package main

func main() {
	log := New(&Config{
		IsProduction: true,
		MaxSize:      100,
		MaxBackups:   7,
		MaxAge:       30,
		Compress:     true,
		IsRotate:     false,
		DirName:      "./logs",
		Filename:     "app.log",
	})

	log.Info("xxx", "Application started")
}

```