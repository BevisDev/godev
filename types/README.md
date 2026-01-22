# Types Package (`types`)

The `types` package provides common type definitions and enums used across GoDev packages.

---

## Overview

This package centralizes shared type definitions to ensure type consistency and avoid duplication across packages.

---

## Types

### Database Types

```go
type DBType int

const (
	SqlServer DBType = iota + 1
	Postgres
	Oracle
	MySQL
)
```

### Other Types

Additional type definitions as needed by the codebase.

---

## Usage

```go
import "github.com/BevisDev/godev/types"

// Use database types
dbType := types.Postgres
```

---

## Notes

- Types are designed to be shared across packages
- Enums use iota for automatic numbering
- Type definitions follow Go naming conventions
