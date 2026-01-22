# redis Package

The `redis` package provides a convenient Redis client for Go, supporting basic operations such
as `GET`, `SET`, `MSET`, `DELETE`, `SCAN`, `EXISTS`, as well as pub/sub features.  
It includes connection pooling, timeout management, and automatic JSON serialization/deserialization.

---

## Features

- Connect to [Redis](https://github.com/redis/go-redis) with host, port, password, database index, pool size, and
  timeout configuration.
- Basic CRUD operations:
    - `Set`, `Setx` (set without expiration)
    - `SetMany`, `SetManyx` (batch set)
    - `Get`, `GetString`, `GetMany`, `GetByPrefix`
    - `Delete`, `Exists`
- Automatic JSON serialization for complex data types (maps, slices, structs, pointers).
- Supports Redis Pub/Sub:
    - `Publish` messages to a channel
    - `Subscribe` to a channel with a message handler
- Context-based timeouts for all operations.

---

## Structure

### `Config`

Configuration struct for connecting to Redis:

| Field        | Type            | Description                                |
|--------------|-----------------|--------------------------------------------|
| `Host`       | `string`        | Redis server hostname or IP.               |
| `Port`       | `int`           | Redis server port.                         |
| `Password`   | `string`        | Password for authentication (if required). |
| `DB`         | `int`           | Redis database index (default 0).          |
| `PoolSize`   | `int`           | Maximum number of connections in the pool. |
| `Timeout`    | `time.Duration` | Timeout for Redis operations.              |

### `Cache`

Main struct for Redis operations (created via `New()`):

| Method        | Description                                          |
|---------------|------------------------------------------------------|
| `GetClient()` | Get underlying Redis client                          |
| `Ping(ctx)`   | Ping Redis server                                    |
| `Close()`     | Close the Redis client connection                    |

### Chain Operations

Chain-based API for type-safe operations:

| Method | Description |
|--------|-------------|
| `With[T](cache *Cache) ChainExec[T]` | Start a chain operation for type T |
| `Key(key string)` | Set the Redis key |
| `Value(value T)` | Set the value to store |
| `Expire(ttl int, unit string)` | Set expiration time |
| `Set(ctx)` | Execute SET operation |
| `Get(ctx)` | Execute GET operation |
| `Delete(ctx)` | Execute DELETE operation |
| `Exists(ctx)` | Check if key exists |

### List Operations

| Method | Description |
|--------|-------------|
| `ListWith[T](cache *Cache) ListExec[T]` | Start list operation |
| `LPush(ctx)`, `RPush(ctx)` | Push to list |
| `LPop(ctx)`, `RPop(ctx)` | Pop from list |
| `LRange(ctx)` | Get list range |

### Set Operations

| Method | Description |
|--------|-------------|
| `SetWith[T](cache *Cache) SetExec[T]` | Start set operation |
| `SAdd(ctx)` | Add to set |
| `SMembers(ctx)` | Get set members |
| `SIsMember(ctx)` | Check membership |

---

```go
package main

import (
	"context"
	"fmt"
	"github.com/BevisDev/godev/redis"
	"log"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func main() {
	ctx := context.Background()

	cache, err := redis.New(&redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	})
	if err != nil {
		log.Fatal("Redis init failed:", err)
	}
	defer cache.Close()

	user := &User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	err = redis.With[User](cache).
		Key("user:1").
		Value(user).
		Expire(3600, "sec").
		Set(ctx)
	if err != nil {
		log.Fatal(err)
	}

	retrieved, err := redis.With[User](cache).
		Key("user:1").
		Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Retrieved user: %+v\n", retrieved)
}

```