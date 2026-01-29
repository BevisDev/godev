# RabbitMQ Queue Management

Production-ready Go package for managing RabbitMQ queues, exchanges, and bindings with comprehensive error handling and flexible configuration.

## Features

- ✅ **Simple Queue Declaration** - Quick setup with sensible defaults
- ✅ **Advanced Configuration** - Full control over queue/exchange parameters
- ✅ **Dead Letter Exchange (DLX)** - Built-in support for failed message handling
- ✅ **Multiple Exchange Types** - Direct, Topic, and Fanout routing
- ✅ **Queue Limits** - TTL, max length, and size constraints
- ✅ **Error Handling** - Comprehensive validation and error messages
- ✅ **Type Safety** - Strongly typed exchange types with validation

## Table of Contents

- [Quick Start](#quick-start)
- [Exchange Types](#exchange-types)
- [Queue Configuration](#queue-configuration)
- [Examples](#examples)
    - [Simple Queues](#1-simple-queues)
    - [Dead Letter Exchange](#2-dead-letter-exchange-dlx)
    - [Topic Exchange](#3-topic-exchange)
    - [Fanout Exchange](#4-fanout-exchange)
    - [Queue Limits](#5-queue-limits)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)

## Quick Start

```go
package main

import (
    "log"
    "your-project/rabbitmq"
)

func main() {
    // Initialize RabbitMQ client
    cfg := &rabbitmq.Config{
        Host:     "localhost",
        Port:     5672,
        Username: "guest",
        Password: "guest",
        Vhost:    "/",
    }
    
    mq, err := rabbitmq.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer mq.Close()
    
    // Declare simple queues
    queue := mq.GetQueue()
    err = queue.Def("orders", "payments", "notifications")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Exchange Types

### Direct Exchange
Routes messages based on **exact routing key match**.

```go
// Routing key: "order.created"
// Matches: "order.created" ✓
// Does NOT match: "order.*" ✗ (patterns not supported)
```

**Use case**: Point-to-point message routing

### Topic Exchange
Routes messages based on **routing key patterns**.

```go
// Routing key: "order.created.email"
// Matches: "order.*.email" ✓
// Matches: "order.#" ✓
// Matches: "#" ✓

// Pattern symbols:
// * = exactly one word
// # = zero or more words
```

**Use case**: Flexible event-based routing (most common)

### Fanout Exchange
**Broadcasts** to all bound queues, ignoring routing key.

```go
// All messages go to ALL bound queues
```

**Use case**: Publishing events to multiple subscribers

## Queue Configuration

### Queue Arguments

```go
const (
    MessageTTL           = "x-message-ttl"             // Time-to-live (ms)
    DeadLetterExchange   = "x-dead-letter-exchange"    // DLX name
    DeadLetterRoutingKey = "x-dead-letter-routing-key" // DLX routing key
    MaxLength            = "x-max-length"              // Max message count
    MaxLengthBytes       = "x-max-length-bytes"        // Max queue size (bytes)
)
```

### QueueSpec Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Name` | `string` | Queue name (required) | - |
| `Durable` | `bool` | Survive broker restart | `false` |
| `AutoDelete` | `bool` | Delete when no consumers | `false` |
| `Exclusive` | `bool` | Single connection only | `false` |
| `Args` | `map[string]interface{}` | Additional arguments | `nil` |

### ExchangeSpec Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Name` | `string` | Exchange name (required) | - |
| `Type` | `ExchangeType` | Direct, Topic, or Fanout | - |
| `Durable` | `bool` | Survive broker restart | `false` |
| `AutoDelete` | `bool` | Delete when no bindings | `false` |
| `Internal` | `bool` | Only exchanges can publish | `false` |
| `Bindings` | `[]BindingSpec` | Queue bindings | `nil` |

## Examples

### 1. Simple Queues

Declare multiple queues with default settings (non-durable).

```go
queue := mq.GetQueue()

err := queue.Def(
    "order.queue",
    "payment.queue",
    "notification.queue",
)
```

### 2. Dead Letter Exchange (DLX)

Handle failed messages by routing them to a dead letter queue.

```go
err := queue.Declare(rabbitmq.Spec{
    Queues: []rabbitmq.QueueSpec{
        {
            Name:    "order.queue",
            Durable: true,
            Args: map[string]interface{}{
                rabbitmq.MessageTTL:           60000, // 60 seconds
                rabbitmq.DeadLetterExchange:   "order.dlx",
                rabbitmq.DeadLetterRoutingKey: "order.failed",
            },
        },
        {
            Name:    "order.dlq", // Dead Letter Queue
            Durable: true,
        },
    },
    Exchanges: []rabbitmq.ExchangeSpec{
        {
            Name:    "order.dlx",
            Type:    rabbitmq.Direct,
            Durable: true,
            Bindings: []rabbitmq.BindingSpec{
                {
                    Queue:      "order.dlq",
                    RoutingKey: "order.failed",
                },
            },
        },
    },
})
```

**Flow**:
1. Message expires in `order.queue` after 60 seconds
2. Routed to `order.dlx` exchange with key `order.failed`
3. Delivered to `order.dlq` for manual inspection

### 3. Topic Exchange

Pattern-based routing for flexible event distribution.

```go
err := queue.Declare(rabbitmq.Spec{
    Queues: []rabbitmq.QueueSpec{
        {Name: "email.queue", Durable: true},
        {Name: "sms.queue", Durable: true},
        {Name: "all-notifications.queue", Durable: true},
    },
    Exchanges: []rabbitmq.ExchangeSpec{
        {
            Name:    "notification.topic",
            Type:    rabbitmq.Topic,
            Durable: true,
            Bindings: []rabbitmq.BindingSpec{
                {Queue: "email.queue", RoutingKey: "*.email"},
                {Queue: "sms.queue", RoutingKey: "*.sms"},
                {Queue: "all-notifications.queue", RoutingKey: "#"},
            },
        },
    },
})
```

**Routing examples**:
- `"order.email"` → `email.queue` + `all-notifications.queue`
- `"user.sms"` → `sms.queue` + `all-notifications.queue`
- `"system.push"` → `all-notifications.queue` only

### 4. Fanout Exchange

Broadcast messages to all consumers.

```go
err := queue.Declare(rabbitmq.Spec{
    Queues: []rabbitmq.QueueSpec{
        {Name: "analytics.queue", Durable: true},
        {Name: "logging.queue", Durable: true},
        {Name: "monitoring.queue", Durable: true},
    },
    Exchanges: []rabbitmq.ExchangeSpec{
        {
            Name:    "events.fanout",
            Type:    rabbitmq.Fanout,
            Durable: true,
            Bindings: []rabbitmq.BindingSpec{
                {Queue: "analytics.queue"},  // routing key ignored
                {Queue: "logging.queue"},
                {Queue: "monitoring.queue"},
            },
        },
    },
})
```

**Result**: Every message published to `events.fanout` reaches all 3 queues.

### 5. Queue Limits

Prevent unbounded queue growth.

```go
err := queue.Declare(rabbitmq.Spec{
    Queues: []rabbitmq.QueueSpec{
        {
            Name:    "limited.queue",
            Durable: true,
            Args: map[string]interface{}{
                rabbitmq.MaxLength:      1000,      // Max 1000 messages
                rabbitmq.MaxLengthBytes: 10485760,  // Max 10MB
            },
        },
    },
})
```

**Behavior**: When limits are reached, oldest messages are dropped (or sent to DLX if configured).

## API Reference

### Queue Methods

#### `Def(names ...string) error`
Declare one or more simple queues with default settings.

```go
err := queue.Def("queue1", "queue2", "queue3")
```

#### `Declare(spec Spec) error`
Declare queues and exchanges with full configuration.

```go
err := queue.Declare(rabbitmq.Spec{
    Queues:    []rabbitmq.QueueSpec{...},
    Exchanges: []rabbitmq.ExchangeSpec{...},
})
```

#### `Delete(name string, ifUnused, ifEmpty bool) error`
Delete a queue.

```go
// Delete only if no consumers and queue is empty
err := queue.Delete("old.queue", true, true)

// Force delete
err := queue.Delete("old.queue", false, false)
```

#### `Purge(name string) error`
Remove all messages from a queue (queue remains).

```go
err := queue.Purge("queue.name")
```

### Spec Types

#### `Spec`
```go
type Spec struct {
    Queues    []QueueSpec
    Exchanges []ExchangeSpec
}
```

#### `QueueSpec`
```go
type QueueSpec struct {
    Name       string
    Durable    bool
    AutoDelete bool
    Exclusive  bool
    Args       map[string]interface{}
}
```

#### `ExchangeSpec`
```go
type ExchangeSpec struct {
    Name       string
    Type       ExchangeType
    Durable    bool
    AutoDelete bool
    Internal   bool
    Bindings   []BindingSpec
}
```

#### `BindingSpec`
```go
type BindingSpec struct {
    Queue      string
    RoutingKey string
    Args       map[string]interface{}
}
```

## Best Practices

### 1. Always Use Durable Queues in Production

```go
// ❌ BAD - Queue lost on broker restart
queue.Def("orders")

// ✅ GOOD - Queue survives restart
queue.Declare(rabbitmq.Spec{
    Queues: []rabbitmq.QueueSpec{
        {Name: "orders", Durable: true},
    },
})
```

### 2. Implement Dead Letter Exchanges

```go
// ✅ GOOD - Failed messages can be inspected and retried
Args: map[string]interface{}{
    rabbitmq.DeadLetterExchange:   "app.dlx",
    rabbitmq.DeadLetterRoutingKey: "failed",
}
```

### 3. Set Message TTL

```go
// ✅ GOOD - Prevent message backlog
Args: map[string]interface{}{
    rabbitmq.MessageTTL: 300000, // 5 minutes
}
```

### 4. Use Queue Limits

```go
// ✅ GOOD - Prevent memory exhaustion
Args: map[string]interface{}{
    rabbitmq.MaxLength:      10000,
    rabbitmq.MaxLengthBytes: 104857600, // 100MB
}
```

### 5. Choose the Right Exchange Type

| Use Case | Exchange Type | Example |
|----------|---------------|---------|
| Single target queue | Direct | Order processing |
| Pattern-based routing | Topic | Notification system |
| Multiple identical consumers | Fanout | Logging/analytics |

### 6. Naming Conventions

```go
// ✅ GOOD - Clear, hierarchical names
"order.created.email"      // Topic routing key
"order.queue"              // Queue name
"notification.topic"       // Exchange name
"order.dlx"                // Dead letter exchange
"order.dlq"                // Dead letter queue
```

### 7. Declare Topology on Startup

```go
func initRabbitMQ() error {
    mq, err := rabbitmq.New(cfg)
    if err != nil {
        return err
    }
    
    // Declare all queues/exchanges at startup
    return mq.GetQueue().Declare(rabbitmq.Spec{
        Queues:    getQueueSpecs(),
        Exchanges: getExchangeSpecs(),
    })
}
```

## Error Handling

All methods return descriptive errors:

```go
err := queue.Def("")
// Error: [queue] at least one queue name is required

err := queue.Delete("nonexistent", true, true)
// Error: delete queue 'nonexistent': NOT_FOUND - no queue 'nonexistent'

err := queue.Purge("blocked.queue")
// Error: [queue] purge queue 'blocked.queue': PRECONDITION_FAILED
```

## Common Patterns

### Microservice Pattern
```go
// Each service declares its own queues
type OrderService struct {
    mq *rabbitmq.RabbitMQ
}

func (s *OrderService) Init() error {
    return s.mq.GetQueue().Declare(rabbitmq.Spec{
        Queues: []rabbitmq.QueueSpec{
            {Name: "order.process", Durable: true},
            {Name: "order.notify", Durable: true},
        },
    })
}
```

### Event-Driven Pattern
```go
// Central event exchange, multiple subscribers
queue.Declare(rabbitmq.Spec{
    Exchanges: []rabbitmq.ExchangeSpec{
        {
            Name:    "events.topic",
            Type:    rabbitmq.Topic,
            Durable: true,
            Bindings: []rabbitmq.BindingSpec{
                {Queue: "analytics.queue", RoutingKey: "#"},
                {Queue: "email.queue", RoutingKey: "user.#"},
                {Queue: "audit.queue", RoutingKey: "*.created"},
            },
        },
    },
})
```

### Retry Pattern
```go
// Main queue with retry via DLX
queue.Declare(rabbitmq.Spec{
    Queues: []rabbitmq.QueueSpec{
        {
            Name:    "order.queue",
            Durable: true,
            Args: map[string]interface{}{
                rabbitmq.DeadLetterExchange: "order.retry",
            },
        },
        {
            Name:    "order.retry.queue",
            Durable: true,
            Args: map[string]interface{}{
                rabbitmq.MessageTTL:         30000, // Retry after 30s
                rabbitmq.DeadLetterExchange: "order.main",
            },
        },
    },
})
```

---

## License

See LICENSE file for details.

## Contributing

Contributions welcome! Please open an issue or submit a pull request.