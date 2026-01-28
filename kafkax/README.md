# kafkax - Quick Start Guide

## 📦 Package Structure

```
type Kafka struct {
    cfg      *Config
    producer *Producer
    consumer *Consumer
}

func New(cfg *Config) (*Kafka, error)
```

Single struct `Kafka` chứa cả Producer và Consumer, khởi tạo thông qua `New()`.

## 🚀 Installation

```bash
go get github.com/segmentio/kafka-go
```

Copy package `kafkax` vào project của bạn:

```bash
cp -r kafkax/ /your/project/pkg/
```

## 📝 Quick Examples

### 1️⃣ Producer Only (Chỉ gửi messages)

```go
package main

import (
	"context"
	"your-project/pkg/kafkax"
)

func main() {
	// Create config
	cfg := kafkax.DefaultConfig([]string{"localhost:9092"})

	// Create Kafka client
	kafka, err := kafkax.New(cfg)
	if err != nil {
		panic(err)
	}
	defer kafka.Close()

	ctx := context.Background()

	// Send message
	kafka.Send(ctx, &kafkax.Message{
		Topic: "orders",
		Key:   []byte("order-1"),
		Value: []byte("order data"),
	})

	// Send JSON
	kafka.SendJSON(ctx, "orders", "order-2", map[string]interface{}{
		"order_id": "order-2",
		"amount":   99.99,
	})
}
```

### 2️⃣ Consumer Only (Chỉ nhận messages)

```go
package main

import (
	"context"
	"fmt"
	"your-project/pkg/kafkax"
)

func main() {
	cfg := kafkax.DefaultConfig([]string{"localhost:9092"})

	// Set consumer config để enable consumer
	cfg.Consumer.GroupID = "order-processor"
	cfg.Consumer.Topics = []string{"orders"}
	cfg.Consumer.AutoCommit = false

	kafka, err := kafkax.New(cfg)
	if err != nil {
		panic(err)
	}
	defer kafka.Close()

	ctx := context.Background()

	// Consume messages
	kafka.Consume(ctx, func(ctx context.Context, msg *kafkax.ConsumedMessage) error {
		fmt.Printf("Received: %s\n", string(msg.Value))

		// Process message here

		return nil // Return nil to commit
	})
}
```

### 3️⃣ Both Producer & Consumer

```go
package main

import (
	"context"
	"your-project/pkg/kafkax"
)

func main() {
	cfg := kafkax.DefaultConfig([]string{"localhost:9092"})

	// Enable consumer
	cfg.Consumer.GroupID = "payment-processor"
	cfg.Consumer.Topics = []string{"payment-requests"}

	kafka, err := kafkax.New(cfg)
	if err != nil {
		panic(err)
	}
	defer kafka.Close()

	ctx := context.Background()

	// Consume requests và produce results
	kafka.Consume(ctx, func(ctx context.Context, msg *kafkax.ConsumedMessage) error {
		// Process payment
		result := processPayment(msg)

		// Send result to another topic
		return kafka.SendJSON(ctx, "payment-results", "key", result)
	})
}
```

## 🎯 Key Concepts

### Config Structure

```go
type Config struct {
    Brokers  []string       // Kafka brokers
    Producer ProducerConfig // Producer settings
    Consumer ConsumerConfig  // Consumer settings
}
```

### Initialization Rules

1. **Producer**: Luôn được khởi tạo (lightweight nếu không dùng)
2. **Consumer**: Chỉ khởi tạo khi set `Consumer.GroupID` và `Consumer.Topics`

```go
cfg := kafkax.DefaultConfig(brokers)

// Producer only (consumer = nil)
kafka, _ := kafkax.New(cfg)

// Both producer and consumer
cfg.Consumer.GroupID = "my-group"
cfg.Consumer.Topics = []string{"topic1"}
kafka, _ := kafkax.New(cfg)
```

### Check Initialization

```go
kafka.HasProducer() // true
kafka.HasConsumer() // true/false (depends on config)
```

## 🔧 Common Operations

### Send Messages

```go
// Simple send
kafka.Send(ctx, &kafkax.Message{
Topic: "orders",
Key:   []byte("key"),
Value: []byte("value"),
})

// Send JSON
kafka.SendJSON(ctx, "orders", "key", orderObject)

// Send batch
messages := []*kafkax.Message{ /* ... */ }
kafka.SendBatch(ctx, messages)

// With headers
producer, _ := kafka.Producer()
producer.SendWithHeaders(ctx, "orders", key, value, map[string]string{
"trace-id": "abc",
})
```

### Consume Messages

```go
// Basic consume
kafka.Consume(ctx, func (ctx context.Context, msg *kafkax.ConsumedMessage) error {
// Process message
return nil // Commit on success
})

// With retry
kafka.ConsumeWithRetry(ctx, handler, 3, time.Second)

// Read single message
consumer, _ := kafka.Consumer()
msg, _ := consumer.ReadMessage(ctx)

// Manual commit
consumer.CommitMessage(ctx, msg)
```

### Access Producer/Consumer Directly

```go
// Get producer
producer, err := kafka.Producer()
if err != nil {
// Handle error (not initialized)
}

// Get consumer
consumer, err := kafka.Consumer()
if err != nil {
// Handle error (not initialized)
}

// Use directly
producer.Send(ctx, msg)
consumer.ReadMessage(ctx)
```

## ⚙️ Configuration Examples

### High Throughput Producer

```go
cfg := kafkax.DefaultConfig(brokers)
cfg.Producer.BatchSize = 1000
cfg.Producer.BatchTimeout = 10 * time.Millisecond
cfg.Producer.Compression = compress.Snappy
cfg.Producer.Async = true
```

### Reliable Consumer

```go
cfg := kafkax.DefaultConfig(brokers)
cfg.Consumer.GroupID = "my-group"
cfg.Consumer.Topics = []string{"orders"}
cfg.Consumer.AutoCommit = false // Manual commit
cfg.Consumer.StartOffset = kafka.FirstOffset
cfg.Consumer.CommitInterval = 1 * time.Second
```

## 🛠️ Development

### Run Local Kafka

```bash
docker-compose up -d      # Start Kafka
docker-compose down -v    # Stop Kafka
```

Access Kafka UI: http://localhost:8080

### Run Tests

```bash
make test           # All tests
make test-short     # Unit tests only
make bench          # Benchmarks
```

### Run Example

```bash
cd examples
go run main.go
```

## ✅ Best Practices

1. **Always defer Close()**
   ```go
   kafka, _ := kafkax.New(cfg)
   defer kafka.Close()
   ```

2. **Use context for timeout**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   ```

3. **Manual commit for reliability**
   ```go
   cfg.Consumer.AutoCommit = false
   ```

4. **Handle errors properly**
   ```go
   kafka.Consume(ctx, func(ctx context.Context, msg *kafkax.ConsumedMessage) error {
       if err := process(msg); err != nil {
           return err // Don't commit on error
       }
       return nil // Commit on success
   })
   ```

5. **Graceful shutdown**
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()
   
   sigChan := make(chan os.Signal, 1)
   signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
   go func() {
       <-sigChan
       cancel()
   }()
   
   kafka.Consume(ctx, handler)
   ```

## 📊 Monitoring

```go
// Producer stats
producer, _ := kafka.Producer()
stats := producer.Stats()
fmt.Printf("Messages: %d, Errors: %d\n", stats.Messages, stats.Errors)

// Consumer lag
consumer, _ := kafka.Consumer()
lag := consumer.Lag()
fmt.Printf("Lag: %d\n", lag)
```

## 🚨 Error Handling

```go
// Config validation error
kafka, err := kafkax.New(cfg)
if err != nil {
// Handle invalid config
}

// Producer not initialized
producer, err := kafka.Producer()
if err == kafkax.ErrProducerNotInitialized {
// Producer not available
}

// Consumer not initialized (GroupID/Topics not set)
consumer, err := kafka.Consumer()
if err == kafkax.ErrConsumerNotInitialized {
// Consumer not available
}

// Message errors
err = kafka.Send(ctx, &kafkax.Message{Topic: ""})
if err == kafkax.ErrEmptyTopic {
// Invalid message
}
```

## 📚 More Examples

Check `examples/main.go` for:

- ✅ Producer only example
- ✅ Consumer only example
- ✅ Both producer and consumer
- ✅ Full featured with retry
- ✅ Direct producer/consumer access

---

**Happy coding! 🚀**