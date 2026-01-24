package kafkax

type kaka struct {
	producer Producer
	consumer Consumer
}

func New(cf *Config) (Client, error) {
	var (
		c   = new(kaka)
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

func (k *kaka) Close() {
	if k.producer != nil {
		k.producer.Close()
	}
	if k.consumer != nil {
		k.consumer.Close()
	}
}

func (k *kaka) GetProducer() Producer {
	return k.producer
}

func (k *kaka) GetConsumer() Consumer {
	return k.consumer
}
