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

| Field        | Description                                |
|--------------|--------------------------------------------|
| `Host`       | Redis server hostname or IP.               |
| `Port`       | Redis server port.                         |
| `Password`   | Password for authentication (if required). |
| `DB`         | Redis database index (default 0).          |
| `PoolSize`   | Maximum number of connections in the pool. |
| `TimeoutSec` | Timeout for Redis operations in seconds.   |

### `RedisCache`

Main struct for Redis operations:

| Method        | Description                                          |
|---------------|------------------------------------------------------|
| `Set`         | Set a key                                            |
| `SetMany`     | Set multiple keys at once.                           |
| `Get`         | Get a key and unmarshal into a struct or basic type. |
| `GetMany`     | Get multiple keys at once.                           |
| `GetByPrefix` | Get all keys and values matching a prefix.           |
| `Delete`      | Delete a key.                                        |
| `Exists`      | Check if a key exists.                               |
| `Publish`     | Publish a message to a Redis channel.                |
| `Subscribe`   | Subscribe to a Redis channel and handle messages.    |
| `Close`       | Close the Redis client connection.                   |

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