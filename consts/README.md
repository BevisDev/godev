# Constants Package (`consts`)

The `consts` package provides common constants used across GoDev packages, including content types, file extensions, regex patterns, and HTTP-related constants.

---

## Overview

This package centralizes commonly used constants to ensure consistency across the codebase and avoid magic strings.

---

## Constants

### Content Types (`content_type.go`)

Common HTTP content types:

```go
const (
	ApplicationJSON       = "application/json"
	ApplicationXML        = "application/xml"
	ApplicationFormURLEncoded = "application/x-www-form-urlencoded"
	ApplicationOctetStream = "application/octet-stream"
	ApplicationPDF        = "application/pdf"
	ApplicationMSWord     = "application/msword"
	ApplicationZip        = "application/zip"
	ApplicationX7z        = "application/x-7z-compressed"
	ApplicationXZip       = "application/x-zip-compressed"
	MultipartFormData     = "multipart/form-data"
	TextPlain             = "text/plain"
	TextHTML              = "text/html"
	TextCSS               = "text/css"
	TextJavaScript        = "text/javascript"
	ImageJPEG             = "image/jpeg"
	ImagePNG              = "image/png"
	ImageGIF              = "image/gif"
	VideoMP4              = "video/mp4"
	AudioMPEG             = "audio/mpeg"
)
```

### File Extensions (`extension.go`)

Common file extensions:

```go
const (
	ExtJSON = ".json"
	ExtYAML = ".yaml"
	ExtYML  = ".yml"
	ExtXML  = ".xml"
	ExtCSV  = ".csv"
	ExtTXT  = ".txt"
	ExtPDF  = ".pdf"
	ExtZIP  = ".zip"
	// ... more extensions
)
```

### Regex Patterns (`pattern.go`)

Common regex patterns for validation:

```go
const (
	Email       = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	PhoneNumber = "^[0-9]{10,11}$"
	// ... more patterns
)
```

### HTTP Constants (`consts.go`)

HTTP-related constants:

```go
const (
	RID     = "rid"      // Request ID key
	Status  = "status"   // HTTP status
	Duration = "duration" // Request duration
	Header  = "header"   // HTTP header
	Body    = "body"     // Request/response body
	URL     = "url"      // Request URL
	Method  = "method"   // HTTP method
	Query   = "query"    // Query parameters
	Time    = "time"     // Timestamp
)
```

---

## Usage

```go
import "github.com/BevisDev/godev/consts"

// Use content types
contentType := consts.ApplicationJSON

// Use HTTP constants
headerKey := consts.RID

// Use patterns
if matched, _ := regexp.MatchString(consts.Email, email); matched {
	// Valid email
}
```

---

## Notes

- Constants are exported and can be used across packages
- Content types follow IANA standards
- Patterns are optimized for common use cases
- HTTP constants are used by logging and middleware packages
