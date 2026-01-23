package kafkax

import (
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type producer struct {
	*Config
	producer *kafka.Producer
	events   chan kafka.Event
}

func NewProducer(cf *Config) (Producer, error) {
	cfg := cf.ProducerConfig
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		clientId:          cf.ClientId,
		bootstrapServers:  cf.BootstrapServers,
		retries:           cfg.Retries,
		acks:              cfg.Acks,
		enableIdempotence: cfg.EnableIdempotence,
		messageMaxBytes:   cfg.MessageMaxBytes,
		requestTimeoutMs:  cfg.RequestTimeoutMs,
		deliveryTimeoutMs: cfg.DeliveryTimeoutMs,
	})
	if err != nil {
		return nil, err
	}

	prod := &producer{
		Config:   cf,
		producer: p,
		events:   make(chan kafka.Event, 100),
	}
	go prod.deliveryHandler()

	return prod, nil
}

func (p *producer) Close() {
	if p.producer != nil {
		p.producer.Flush(5000)
		close(p.events)
		p.producer.Close()
	}
}

func (p *producer) deliveryHandler() {
	for e := range p.events {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				log.Printf("Delivery failed: %v", ev.TopicPartition.Error)
			}
		case kafka.Error:
			log.Printf("Producer error: %v", ev)
			if ev.IsFatal() || ev.Code() == kafka.ErrAllBrokersDown {
				log.Println("All brokers down - closing producer")
				p.Close()
				return
			}
		}
	}
}

func (p *producer) Produce(
	id, topic string,
	key, value []byte,
) error {
	return p.producer.Produce(
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
		p.events,
	)
}
