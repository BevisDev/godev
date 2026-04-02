# RabbitMQ Consumer

Consumer side uses a small **manager type `CM`** (constructed when `rabbitmq.New` runs with consumer enabled). You register one or more `Consumer` values (queue + handler), then call **`Start(ctx)`** so each active consumer runs in its own goroutine with reconnect-style error handling.

## Obtaining the consumer manager

```go
mq, err := rabbitmq.New(cfg)
if err != nil { /* ... */ }
defer mq.Close()

cm := mq.Consumer()
if cm == nil {
    // nil if rabbitmq.WithProducerOnly() was used
}
```

Disable consuming when the process only publishes:

```go
mq, err := rabbitmq.New(cfg, rabbitmq.WithProducerOnly())
```

## Registration and lifecycle

| Step | API | Notes |
|------|-----|------|
| Register | `cm.Register(&Consumer{ ... })` | Skips entries with empty `Queue` or nil `Handler`. Same queue name **replaces** the previous registration. |
| Run | `go cm.Start(ctx)` or `cm.Start(ctx)` | **Blocks** until `ctx` is cancelled; starts one goroutine per consumer where `IsOn == true`. |
| Shutdown | `mq.Close()` → `cm.Close()` | `Close` waits up to **5s** for worker goroutines to finish after `Start` returns. |

Typical pattern (e.g. in bootstrap):

```go
cm.Register(
    &rabbitmq.Consumer{ Queue: "orders", Handler: h, IsOn: true },
)
go cm.Start(appCtx)
```

## `Consumer` fields

| Field | Role |
|-------|------|
| `IsOn` | If false, this queue is **not** started. |
| `Queue` | AMQP queue name. `Consume` also calls `CreateQueues` for that name before subscribing. |
| `Handler` | Your `Handler` implementation (see below). |
| `PrefetchCount` | `Channel.Qos(prefetch, 0, false)`. If **≤ 0**, uses manager default **1**. |
| `WorkerPool` | Goroutines draining an internal buffered channel of deliveries. If **≤ 0**, manager default **10**. |
| `MaxConsecutiveErrors` | After this many **consecutive** failures from `Consume` (channel setup / delivery loop), that consumer loop exits. If **≤ 0**, default **10**. |
| `RetryDelay` | Sleep before retrying after an error. If **≤ 0**, default **5s**. |

Manager-level defaults (`prefetch`, `workerPool`, etc.) are fixed inside `newCM` unless overridden per `Consumer` as above.

## `Handler` and acknowledgments

```go
type Handler interface {
    Handle(ctx context.Context, msg *MsgHandler) error
}
```

- **`ConsumeWithContext`** is called with **manual ack** (`autoAck == false`). Ack/nack/requeue are your responsibility unless `WithAutoCommit` is enabled on the **MQ client** (see below).
- **Context** passed to `Handle` is built with `utils.NewCtx()` and **`consts.RID`** set from the delivery’s `CorrelationId` (via `utils.SetValueCtx`).

### `WithAutoCommit()` (MQ option)

- **`autoCommit == true`**: after `Handle` returns **nil**, the client calls **`Ack(false)`** automatically. On non-nil error, the client **does not** call `Requeue`/`Nack` in the auto-commit path (only the manual path calls `Requeue` on error).
- **`autoCommit == false`** (default): on **success** you should call **`msg.Commit()`** (or `CommitMulti`). On **failure** the manager calls **`Requeue()`** (`Nack(false, true)`).

Doc comments on `Handler` state: *returns nil to ack; error to requeue* — that matches the **manual** ack path; with auto-commit, success still results in an automatic ack.

### Panics

`handleMsg` recovers panics, returns an error, and calls **`Reject()`** (`Reject(false)` — message is discarded, not requeued).

## Message API (`MsgHandler`)

Useful accessors: `QueueName`, `GetBody`, `BodyAs[T]`, `ContentType`, `CorrelationID`, `Timestamp`, `Header(key)`.

Ack helpers: `Commit`, `CommitMulti`, `Requeue`, `RequeueMulti`, `Reject`, `RejectRequeue`.

## Inner pipeline (one consumer)

1. Open a channel; **`CreateQueues(queueName)`** (declare queue if your queue helper defines it that way).
2. **`Qos`** with effective prefetch.
3. **`ConsumeWithContext(ctx, queue, ...)`** — not exclusive, not no-wait; consumer tag auto.
4. Delivery loop: push each `amqp.Delivery` into a buffered `jobs` channel (`capacity == workerCount`); workers call `processMsg`.
5. Cancel or closed delivery channel ends the loop; workers drain `jobs` in `defer`.

## Errors and retries

- **Transient broker/channel codes** `504`, `320`, `501` are logged as connection-style; after `RetryDelay`, the **Run** loop calls `Consume` again.
- Other errors: still increment consecutive error count, sleep `RetryDelay`, retry until `MaxConsecutiveErrors`.
- When context is done, `Consume` returns **nil** and the loop exits cleanly.

There is **no separate “consume timeout”** option on the client; cancellation is entirely via **`ctx`**.

## Options related to consuming

| Option | Effect |
|--------|--------|
| `WithAutoCommit()` | Auto-ack after successful `Handle` (see semantics above). |
| `WithProducerOnly()` | No `CM`; `Consumer()` is nil. |
| `WithConsumerOnly()` | Producer nil; consumer available. |
| `WithReconnectMaxRetries` | Applies to **connection** reconnect in `MQ`, not per-message. |

## See also

- [PRODUCER.md](./PRODUCER.md) — publishing.
- [QUEUE_DOC.md](./QUEUE_DOC.md) — declaring queues, exchanges, bindings (must match how publishers route into your queues).
