# RabbitMQ Package

`rabbitmq` is a simple Go client for RabbitMQ with automatic reconnection, queue management, and message handling.

## Features

- **Connect**: Establishes connection to RabbitMQ server using config.
- **GetChannel**: Returns a new channel for publishing or consuming.
- **DeclareQueue**: Ensures a queue exists before use.
- **Publish**: Sends messages (string, []byte, int, bool, JSON structs) with optional context header `Xstate`.
- **Consume**: Starts consuming messages with a handler callback; automatically extracts `Xstate` from headers.
- Thread-safe and handles reconnection automatically.

## Usage

```go
cfg := &rabbitmq.Config{
    Host:     "localhost",
    Port:     5672,
    Username: "guest",
    Password: "guest",
    VHost:    "/",
}

client, err := rabbitmq.NewRabbitMQ(cfg)
if err != nil {
    log.Fatalf("failed to connect RabbitMQ: %v", err)
}

// Declare queue
client.DeclareQueue("my-queue")

ctx := context.Background()

// Publish message
client.Publish(ctx, "my-queue", map[string]string{"hello": "world"})

// Consume messages
client.Consume(ctx, "my-queue", func(ctx context.Context, msg amqp.Delivery) {
    log.Println("Received:", string(msg.Body))
    msg.Ack(false)
})
