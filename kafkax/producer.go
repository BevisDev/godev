package kafkax

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer struct {
	cf       *Config
	producer *kafka.Producer
	events   chan kafka.Event
	mu       sync.Mutex
	closed   bool
}

func NewProducer(cf *Config) (*Producer, error) {
	cfg := cf.ProducerConfig
	configMap := kafka.ConfigMap{
		clientId:         cf.ClientId,
		bootstrapServers: cf.BootstrapServers,
	}

	if cfg != nil {
		if cfg.Retries > 0 {
			configMap[retries] = cfg.Retries
		}
		if cfg.Acks > 0 {
			configMap[acks] = cfg.Acks
		}
		if cfg.EnableIdempotence {
			configMap[enableIdempotence] = cfg.EnableIdempotence
		}
		if cfg.MessageMaxBytes > 0 {
			configMap[messageMaxBytes] = cfg.MessageMaxBytes
		}
		if cfg.RequestTimeoutMs > 0 {
			configMap[requestTimeoutMs] = cfg.RequestTimeoutMs
		}
		if cfg.DeliveryTimeoutMs > 0 {
			configMap[deliveryTimeoutMs] = cfg.DeliveryTimeoutMs
		}
	}

	p, err := kafka.NewProducer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("[producer] failed %w", err)
	}

	prod := &Producer{
		cf:       cf,
		producer: p,
		events:   make(chan kafka.Event, 100),
	}
	go prod.deliveryHandler()

	return prod, nil
}

func (p *Producer) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true

	if p.producer != nil {
		p.producer.Flush(5000)
		if p.events != nil {
			close(p.events)
		}
		p.producer.Close()
	}
}

func (p *Producer) deliveryHandler() {
	for e := range p.events {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v", ev.TopicPartition.Error)
			}
		case kafka.Error:
			log.Printf("Producer error: %v", ev)
			if ev.IsFatal() || ev.Code() == kafka.ErrAllBrokersDown {
				log.Println("All brokers down - closing Producer")
				p.Close()
				return
			}
		}
	}
}

// Produce sends a message to Kafka with the given topic, key, and value.
func (p *Producer) Produce(
	id, topic string,
	key, value []byte,
) error {
	p.mu.Lock()
	closed := p.closed
	producer := p.producer
	events := p.events
	p.mu.Unlock()

	if closed || producer == nil {
		return fmt.Errorf("[producer] producer is closed")
	}

	return producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   key,
			Value: value,
			Headers: []kafka.Header{
				{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
				{Key: "id", Value: []byte(id)},
			},
		},
		events,
	)
}

// ProduceString sends a message with string key and value.
func (p *Producer) ProduceString(
	id, topic string,
	key, value string,
) error {
	return p.Produce(id, topic, []byte(key), []byte(value))
}

// ProduceJSON serializes the value to JSON and sends it to Kafka.
func (p *Producer) ProduceJSON(
	id, topic string,
	key string,
	value interface{},
) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return p.ProduceString(id, topic, key, string(jsonValue))
}

// ProduceWithHeaders sends a message with custom headers.
func (p *Producer) ProduceWithHeaders(
	id, topic string,
	key, value []byte,
	customHeaders map[string]string,
) error {
	p.mu.Lock()
	closed := p.closed
	producer := p.producer
	events := p.events
	p.mu.Unlock()

	if closed || producer == nil {
		return fmt.Errorf("[producer] producer is closed")
	}

	headers := []kafka.Header{
		{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		{Key: "id", Value: []byte(id)},
	}

	for k, v := range customHeaders {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	return producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:     key,
			Value:   value,
			Headers: headers,
		},
		events,
	)
}

// ProduceToPartition sends a message to a specific partition.
func (p *Producer) ProduceToPartition(
	id, topic string,
	partition int32,
	key, value []byte,
) error {
	p.mu.Lock()
	closed := p.closed
	producer := p.producer
	events := p.events
	p.mu.Unlock()

	if closed || producer == nil {
		return fmt.Errorf("[producer] producer is closed")
	}

	return producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: partition,
			},
			Key:   key,
			Value: value,
			Headers: []kafka.Header{
				{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
				{Key: "id", Value: []byte(id)},
			},
		},
		events,
	)
}

// Flush waits for all pending messages to be delivered.
func (p *Producer) Flush(timeoutMs int) int {
	p.mu.Lock()
	producer := p.producer
	p.mu.Unlock()

	if producer == nil {
		return 0
	}
	return producer.Flush(timeoutMs)
}
