# kafkax – Production Readiness Review

## Overview

Package `kafkax` is a wrapper around `github.com/segmentio/kafka-go` with sensible configuration, a unified producer/consumer API, and support for RID (request ID), retries, and manual commit.

---

## What’s Already Production-Ready

| Area | Details |
|------|---------|
| **Config** | `Validate()`, `DefaultConfig()`, and config is cloned in `New()` so caller mutations do not affect the client |
| **Producer** | Send, SendBatch, SendJSON, SendWithHeaders, Produce (RID), ProduceBatch; mutex and closed checks; Stats, Close, IsClosed |
| **Consumer** | Consume, ConsumeWithRetry, ReadMessage, CommitMessage; manual/auto commit; Lag, Stats, SetOffset; RID from header → context |
| **Errors** | Clear sentinel errors (ErrNoBrokers, ErrProducerClosed, ErrConsumerNotInitialized, etc.) |
| **Graceful** | Consumer exits on `ctx.Done()`; `Close()` shuts down both producer and consumer and logs close errors |
| **Poison message** | ConsumeWithRetry: after retries are exhausted the message is **committed (skipped)** and logged, so the partition is not blocked forever |

---

## Fixes Applied in This Review

1. **kafka.go**
   - Producer init errors are no longer ignored; `New()` returns an error if `newProducer` fails.
   - `Close()`: log errors when closing producer/consumer and set references to `nil` after close.
   - Config is cloned inside `New()`.

2. **producer.go**
   - `Produce` / `ProduceBatch`: added lock, closed and nil writer checks, topic validation; ProduceBatch preserves each message’s headers and adds RID.

3. **consumer.go**
   - `ConsumeWithRetry`: when retries are exhausted, the message is committed (skipped) and logged to avoid an infinite loop on a single failing message.

4. **config.go**
   - Documented that `Idempotent` is not yet applied to the kafka-go Writer (reserved for when the driver supports it).

---

## Further Recommendations for Production

### 1. Logging

- The package currently uses `fmt.Printf` / `log.Printf`. Prefer injecting a logger (e.g. `logger.Logger`) via config or options so logs align with the rest of the app (level, format, tracing).

### 2. Observability

- Producer: `Stats()` (kafka.WriterStats) is available – consider exporting metrics (message count, errors, batch size).
- Consumer: `Stats()` and `Lag()` are available – monitor lag and messages/sec to detect backlogs.

### 3. Consumer: AutoCommit

- `AutoCommit == true`: kafka-go commits according to `CommitInterval`.
- `AutoCommit == false`: commit only after the handler returns successfully. This gives at-least-once delivery; implement idempotent handling in the handler if you need to avoid duplicates.

### 4. Idempotent / Exactly-once

- `Config.Producer.Idempotent` exists but is **not** yet passed to the kafka-go Writer (the driver does not fully support it). When kafka-go supports it, wire it in `newProducer`.

### 5. Consumer SetOffset

- `SetOffset(topic, partition, offset)` currently only calls `reader.SetOffset(offset)`; the topic/partition arguments are unused (the Reader is group-based). To set offset per partition you need a different approach (e.g. a Reader per partition or a low-level API). Consider renaming or documenting this to avoid confusion.

### 6. Tests

- There are no `*_test.go` files yet. Add unit tests for Validate and config clone, and integration tests (e.g. testcontainers or mocks) for basic Send/Consume flows.

### 7. Graceful Shutdown

- The consumer runs in a loop; when `ctx.Done()` (shutdown), `Consume` returns. In main/bootstrap, cancel the context on SIGTERM/SIGINT and then call `Kafka.Close()` so producer and consumer shut down cleanly.

---

## Conclusion

- **Suitable for production** after the changes in this review: correct producer init, thread-safe Produce/ProduceBatch, poison-message handling in ConsumeWithRetry, and safe Close with config clone.
- **Recommended next steps**: inject a logger, monitor Stats/Lag, add tests, and (when needed) clarify or refactor the SetOffset API and enable Idempotent once the driver supports it.
