# RabbitMQ Producer

Publisher API for sending messages with [`amqp091-go`](https://github.com/rabbitmq/amqp091-go). The producer borrows a **new channel per publish**, runs `PublishWithContext`, then closes the channel (see `MQ.WithChannel`).

## Obtaining a producer

```go
cfg := &rabbitmq.Config{ /* Host, Port, Username, Password, VHost */ }

mq, err := rabbitmq.New(cfg) // default: producer + consumer enabled
if err != nil { /* ... */ }
defer mq.Close()

prod := mq.Producer()
if prod == nil {
    // nil if rabbitmq.WithConsumerOnly() was used (producer disabled)
}
```

Disable the producer when the process only consumes:

```go
mq, err := rabbitmq.New(cfg, rabbitmq.WithConsumerOnly())
```

## Public methods

| Method | Routing | Typical broker setup |
|--------|---------|----------------------|
| `Send(ctx, queueName, message, props...)` | **Default exchange**, routing key = **queue name** | Point-to-point: message goes straight to that queue. |
| `PublishEvent(ctx, exchange, routingKey, message, props...)` | Named **exchange** + **routing key** | Direct / topic / custom exchange; queues bind with matching keys. |
| `BroadcastEvent(ctx, exchange, message, props...)` | Named **exchange**, **empty** routing key | **Fanout** exchange: every bound queue receives a copy. |

All three share the same internal pipeline: build `amqp.Publishing` → `Channel.PublishWithContext` with **mandatory** `true` and **immediate** `false`.

## Context and timeouts

There is **no built-in publish timeout** on the client. Cancellation and deadlines come only from `ctx`:

- Use `context.WithTimeout` / `WithDeadline` when you need an upper bound on how long publish may block.
- If `ctx` has no deadline, the call can wait until the context is cancelled or the network/server completes the operation.

## Message body and size

- Payload is serialized with `utils.ToBytes(message)` (supports types your helper allows: e.g. `[]byte`, `string`, JSON-friendly structs, etc.).
- Maximum body size is enforced: **50 000 bytes**. Larger payloads return `ErrMessageTooLarge`.
- **Content-Type** is set to **plain text** by default; if the encoded bytes are valid JSON, **application/json** is used.

## Message properties (optional)

Functional options of type `MsgProperties`:

| Option | Effect |
|--------|--------|
| `WithPersistentMsg()` | `DeliveryMode = Persistent` (survive broker restart if the queue is durable). |
| `WithMessageID(id)` | Sets `MessageId`. |
| `WithCorrelationID(id)` | Sets `CorrelationId`; if omitted, a value from `utils.GetRID(ctx)` is used when present. |
| `WithHeaders(map[string]any)` | AMQP `Headers` table. |
| `WithAppID`, `WithUserID`, `WithReplyTo` | Standard AMQP properties. |
| `WithTimestamp(time.Time)` | If zero, `time.Now()` is used. |
| `WithExpiration(d time.Duration)` | Per-message TTL; value is sent as **milliseconds** (string), as required by RabbitMQ. |

Example:

```go
err := prod.PublishEvent(ctx, "orders.events", "order.created", payload,
    rabbitmq.WithPersistentMsg(),
    rabbitmq.WithHeaders(map[string]any{"version": 1}),
)
```

## Errors

| Error | Meaning |
|-------|---------|
| `ErrMessageTooLarge` | Body exceeds 50 000 bytes after encoding. |
| `ErrNilConfig` / connection errors | From `New` or `WithChannel` / `GetChannel` when the client or connection is invalid. |

Wrap errors from `publish` may include `build message: ...` when serialization fails.

## Examples

### Send to a queue by name (default exchange)

```go
err := prod.Send(ctx, "my-queue", map[string]any{"event": "ping"})
```

### Topic-style event

```go
err := prod.PublishEvent(ctx, "app.topic", "user.signup", eventPayload)
```

### Fanout broadcast

```go
err := prod.BroadcastEvent(ctx, "notifications.fanout", notificationPayload)
```

Ensure exchanges, queues, and bindings are declared (for example via `Queue` / `Spec` in this package) before publishing.

## See also

- [CONSUMER.md](./CONSUMER.md) — consumer manager, handlers, ack modes.
- [QUEUE_DOC.md](./QUEUE_DOC.md) — queues, exchanges, bindings.
- `MQ.Producer()`, `MQ.WithChannel`, `rabbitmq/options.go` — client options.
