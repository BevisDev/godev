# RestClient Package

The `rest` package provides a powerful and flexible REST client for Go, with optional logging support.  
It supports common HTTP methods such as `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`, with JSON or form-data payloads,
query and path parameters, and automatic response unmarshaling into structs.

---

## Features

- Supports HTTP methods: `GET`, `POST`, `POST Form`, `PUT`, `PATCH`, `DELETE`.
- Send data as `JSON` or `application/x-www-form-urlencoded`.
- Handles query parameters (`?key=value`) and path parameters (`/users/:id`).
- Detailed request/response logging, with options to skip logging headers or body.
- Configurable request timeout.
- Automatic unmarshaling of response JSON into Go structs.
- Supports raw byte responses if needed.
- Can skip logging for specific APIs or content types.

---

## Structure

### `Request`

Struct for configuring each HTTP request:

| Field      | Description                                             |
|------------|---------------------------------------------------------|
| `URL`      | API endpoint (e.g., `/users/:id`).                      |
| `Query`    | Map of query parameters (`?key=value`).                 |
| `Params`   | Map of path parameters (`:id`).                         |
| `BodyForm` | Map of form data (`application/x-www-form-urlencoded`). |
| `Header`   | Map of custom headers.                                  |
| `Body`     | Request body (struct will be JSON-encoded).             |
| `Result`   | Pointer to a struct for unmarshaling the response.      |

### `HttpConfig`

Configuration struct for `RestClient`:

| Field             | Description                                         |
|-------------------|-----------------------------------------------------|
| `TimeoutSec`      | Request timeout in seconds.                         |
| `Logger`          | Logger instance (`AppLogger`) for detailed logging. |
| `SkipLogHeader`   | Skip logging headers if `true`.                     |
| `SkipLogAPIs`     | List of API paths to skip logging body.             |
| `SkipContentType` | List of content types to skip logging body.         |

### `RestClient`

Main struct to perform HTTP requests.

Key methods:

- `Get(ctx, req *Request) error`
- `Post(ctx, req *Request) error`
- `PostForm(ctx, req *Request) error`
- `Put(ctx, req *Request) error`
- `Patch(ctx, req *Request) error`
- `Delete(ctx, req *Request) error`

---

```go
package main

import (
	"context"
	"fmt"
	"github.com/BevisDev/godev/rest"
)

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var Client *rest.RestClient

func main() {
	ctx := context.Background()

	Client = rest.New(&rest.HttpConfig{
		TimeoutSec:    10,
		SkipLogHeader: true,
	})

	user, err := rest.NewRequest[UserResponse](Client).
		URL("https://jsonplaceholder.typicode.com/users/:id").
		Params(map[string]string{"id": "1"}).
		GET(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("User: %+v\n", user)
}

```