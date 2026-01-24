package kafkax

type Kafka struct {
	producer *Producer
	consumer *Consumer
}

func New(cf *Config) (*Kafka, error) {
	var (
		c   = new(Kafka)
		err error
	)

	if cf.ProducerConfig != nil {
		c.producer, err = NewProducer(cf)
		if err != nil {
			return nil, err
		}
	}

	if cf.ConsumerConfig != nil {
		c.consumer, err = NewConsumer(cf)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (k *Kafka) Close() {
	if k.producer != nil {
		k.producer.Close()
	}
	if k.consumer != nil {
		k.consumer.Close()
	}
}

func (k *Kafka) GetProducer() *Producer {
	return k.producer
}

func (k *Kafka) GetConsumer() *Consumer {
	return k.consumer
}
