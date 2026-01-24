package kafkax

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

type Client interface {
	Close()
	GetProducer() Producer
	GetConsumer() Consumer
}

type Producer interface {
	Close()
	Produce(id, topic string, key, value []byte) error
	ProduceString(id, topic string, key, value string) error
	ProduceJSON(id, topic string, key string, value interface{}) error
	ProduceWithHeaders(id, topic string, key, value []byte, customHeaders map[string]string) error
	ProduceToPartition(id, topic string, partition int32, key, value []byte) error
	Flush(timeoutMs int) int
}

// Handler defines the interface for handling Kafka messages.
// Implement this interface to process messages.
type Handler interface {
	Handle(msg *kafka.Message) error
}

// Consumer interface simplified to only require service implementation.
type Consumer interface {
	Close()
	// Start begins consuming messages and calls the handler for each message.
	// It blocks until an error occurs or the consumer is closed.
	Start(topics []string, handler Handler) error
	// StartWithTimeout starts consuming with a timeout per message poll.
	StartWithTimeout(topics []string, handler Handler, timeoutMs int) error
}
