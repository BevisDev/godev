package kafkax

type Kafka struct {
	cfg      *Config
	producer *Producer
	consumer *Consumer
}

func New(cf *Config) (*Kafka, error) {
	var (
		err error
		c   = &Kafka{cfg: cf}
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
