# Rest Client

The `rest` package provides a powerful and flexible REST client for Go, with optional logging support.  
It supports common HTTP methods such as `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`, with JSON or form-data payloads,
query and path parameters, and automatic response unmarshaling into structs, and configurable request/response logging.

---

## Features

- Supports HTTP methods:
    - `GET`, `POST`, `POST Form`
    - `PUT`, `PATCH`, `DELETE`
- Path parameters (`/users/:id`) and query parameters
- Generic response handling (`Request[T]`)
- Configurable request timeout
- Detailed request/response logging
- Skip logging by:
    - Header
    - API path
    - Content-Type
- Suitable for:
    - Internal SDKs
    - Service-to-service communication
    - Microservices

---

## Structure

### `Request[T]`

`Request[T]` is a **request builder** (the internal struct is not exposed).  
It is used to configure and execute a single HTTP request with a **type-safe response**.

| Method                          | Description                                     |
|---------------------------------|-------------------------------------------------|
| `URL(string)`                   | API endpoint (e.g. `/users/:id`)                |
| `Query(map[string]string)`      | Query parameters (`?key=value`)                 |
| `PathParams(map[string]string)` | Path parameters (`:id`)                         |
| `Header(map[string]string)`     | Custom HTTP headers                             |
| `Body(any)`                     | Request body (automatically JSON-encoded)       |
| `BodyForm(map[string]string)`   | Form body (`application/x-www-form-urlencoded`) |

The response body is **automatically unmarshaled** into type `T`.

### Client Options

`RestClient` is configured using the **Option Pattern**.  
The internal configuration struct is not exposed.

| Option                                  | Description                                               |
|-----------------------------------------|-----------------------------------------------------------|
| `WithTimeout(time.Duration)`            | Set request timeout                                       |
| `WithLogger(logx.Logger)`               | Enable request/response logging                           |
| `WithSkipHeader()`                      | Skip logging HTTP headers                                 |
| `WithSkipBodyByPaths(...string)`        | Skip logging body for specific API paths                  |
| `WithSkipBodyByContentTypes(...string)` | Skip logging body for specific content types              |
| `WithSkipDefaultContentTypeCheck()`     | Disable the default content-type based body logging check |

---

### `RestClient`

`RestClient` is the main HTTP client.  
It should be created once and reused across the application.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/BevisDev/godev/rest"
)

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var Client *rest.Client

func main() {
	ctx := context.Background()

	Client = rest.New(
		rest.WithTimeout(10 * time.Second),
	)

	user, err := rest.NewRequest[*UserResponse](Client).
		URL("https://jsonplaceholder.typicode.com/users/:id").
		PathParams(map[string]string{
			"id": "1",
		}).
		GET(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("User: %+v\n", user)
}

```