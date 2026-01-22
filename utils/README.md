# Utils Package

The `utils` package provides a collection of utility functions for common Go development tasks, including string manipulation, validation, date/time handling, cryptography, file operations, JSON utilities, and more.

---

## Packages

### Core Utilities (`utils`)

Core utility functions for context management, masking, and general helpers.

**Key Functions:**
- `NewCtx()` - Create context with RID
- `SetValueCtx()` - Set value in context
- `GetRID()` - Get Request ID from context
- `NewCtxTimeout()` - Create context with timeout
- `MaskLeft()`, `MaskRight()`, `MaskCenter()` - String masking utilities
- `MaskEmail()` - Email masking
- `SkipContentType()` - Check if content type should be skipped
- `Parse[T]()` - Type-safe parsing
- `IsContains()`, `IndexOf()` - Slice utilities

**Example:**
```go
import "github.com/BevisDev/godev/utils"

// Create context with RID
ctx := utils.NewCtx()

// Mask email
masked := utils.MaskEmail("john.doe@example.com", 3, 4)
// Result: "john.***@exam****"

// Check if content type should be skipped
shouldSkip := utils.SkipContentType("image/png") // true
```

---

### String Utilities (`utils/str`)

Comprehensive string manipulation and formatting functions.

**Key Functions:**
- `ToString()` - Convert any value to string
- `StartWith()`, `EndWith()` - String prefix/suffix checking
- `ContainsIgnoreCase()` - Case-insensitive contains
- `TrimSpace()` - Trim whitespace
- `ToLower()`, `ToUpper()` - Case conversion
- `Replace()`, `ReplaceAll()` - String replacement
- `Split()`, `Join()` - String splitting and joining

**Example:**
```go
import "github.com/BevisDev/godev/utils/str"

// Convert to string
val := str.ToString(123) // "123"

// Check prefix/suffix
if str.StartWith("hello", "he") {
	// true
}

// Case-insensitive contains
if str.ContainsIgnoreCase("Hello World", "hello") {
	// true
}
```

---

### Validation (`utils/validate`)

Type validation and checking utilities.

**Key Functions:**
- `IsNilOrEmpty()` - Check if value is nil or empty
- `IsNonNilPointer()` - Check if value is a non-nil pointer
- `IsStruct()` - Check if value is a struct
- `IsTimedOut()` - Check if error is a timeout
- `Matches()` - Regex pattern matching
- `IsEmail()` - Email validation
- `IsPhoneNumber()` - Phone number validation

**Example:**
```go
import "github.com/BevisDev/godev/utils/validate"

// Check if pointer
if validate.IsNonNilPointer(&value) {
	// true
}

// Validate email
if validate.IsEmail("user@example.com") {
	// true
}

// Check timeout error
if validate.IsTimedOut(err) {
	// Handle timeout
}
```

---

### Date/Time Utilities (`utils/datetime`)

Comprehensive date and time manipulation utilities.

**Key Functions:**
- `ToString()` - Format time to string
- `Parse()` - Parse string to time
- `Now()` - Get current time (UTC/Local)
- `Add()`, `Sub()` - Time arithmetic
- `Format()` - Custom formatting
- `ToUTC()`, `ToLocal()` - Timezone conversion

**Example:**
```go
import (
	"github.com/BevisDev/godev/utils/datetime"
	"time"
)

// Format time
formatted := datetime.ToString(time.Now(), datetime.DateTimeLayoutMilli)

// Parse time
t, err := datetime.Parse("2024-01-15 10:30:45", datetime.DateTimeLayout)

// Get UTC time
utcTime := datetime.UTCNow()

// Get local time
localTime := datetime.LocalNow()
```

---

### Cryptography (`utils/crypto`)

Encryption and decryption utilities.

**Key Functions:**
- `EncryptAES()` - AES encryption (CTR mode)
- `DecryptAES()` - AES decryption (CTR mode)
- `EncryptRSA()` - RSA encryption
- `DecryptRSA()` - RSA decryption
- `GenerateKey()` - Generate encryption key

**Example:**
```go
import "github.com/BevisDev/godev/utils/crypto"

key := []byte("32-byte-key-for-AES-256-encryption!!")

// Encrypt
encrypted, err := crypto.EncryptAES(key, []byte("secret data"))

// Decrypt
decrypted, err := crypto.DecryptAES(key, encrypted)
```

---

### File Utilities (`utils/filex`)

File and directory operations.

**Key Functions:**
- `ReadFile()` - Read file content
- `WriteFile()` - Write file content
- `Exists()` - Check if file exists
- `CreateDir()` - Create directory
- `ListFiles()` - List files in directory

**Example:**
```go
import "github.com/BevisDev/godev/utils/filex"

// Read file
content, err := filex.ReadFile("config.yaml")

// Write file
err := filex.WriteFile("output.txt", []byte("content"))

// Check existence
if filex.Exists("file.txt") {
	// File exists
}
```

---

### JSON Utilities (`utils/jsonx`)

JSON marshaling and unmarshaling utilities.

**Key Functions:**
- `Marshal()` - Marshal to JSON
- `Unmarshal()` - Unmarshal from JSON
- `Pretty()` - Pretty print JSON

**Example:**
```go
import "github.com/BevisDev/godev/utils/jsonx"

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Marshal
data, err := jsonx.Marshal(User{Name: "Alice", Email: "alice@example.com"})

// Unmarshal
var user User
err := jsonx.Unmarshal(data, &user)
```

---

### Money Utilities (`utils/money`)

Financial calculations and formatting.

**Key Functions:**
- `Format()` - Format money amount
- `Parse()` - Parse money string
- `Add()`, `Sub()`, `Mul()`, `Div()` - Money arithmetic

**Example:**
```go
import "github.com/BevisDev/godev/utils/money"

amount := money.New(1000000) // 1,000,000 (in smallest unit)
formatted := amount.Format("VND") // "1,000,000 VND"
```

---

### Random Utilities (`utils/random`)

Random value generation.

**Key Functions:**
- `NewUUID()` - Generate UUID
- `NewInt()` - Random integer
- `NewFloat()` - Random float
- `Item()` - Random item from slice

**Example:**
```go
import "github.com/BevisDev/godev/utils/random"

// Generate UUID
id := random.NewUUID()

// Random integer
n := random.NewInt(1, 100) // 1 <= n < 100

// Random item
items := []string{"apple", "banana", "cherry"}
item := random.Item(items)
```

---

## Best Practices

1. **Use type-safe functions**: Prefer generic functions like `Parse[T]()` when available
2. **Handle errors**: Always check errors from utility functions
3. **Use appropriate utilities**: Choose the right utility for your use case
4. **Validate inputs**: Use validation utilities before processing
5. **Format consistently**: Use datetime utilities for consistent time formatting

---

## Integration

All utilities are designed to work seamlessly with other GoDev packages:

```go
import (
	"github.com/BevisDev/godev/utils"
	"github.com/BevisDev/godev/utils/str"
	"github.com/BevisDev/godev/utils/validate"
	"github.com/BevisDev/godev/utils/datetime"
)

// Use in your application
ctx := utils.NewCtx()
rid := utils.GetRID(ctx)

if validate.IsEmail(email) {
	// Process email
}

formatted := datetime.ToString(time.Now(), datetime.DateTimeLayout)
```
