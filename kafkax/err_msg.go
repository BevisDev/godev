package kafkax

import (
	"errors"
)

var (
	ErrNilConfig              = errors.New("[kafkax] config is nil")
	ErrNoBrokers              = errors.New("[kafkax] no brokers")
	ErrBothDisabled           = errors.New("[kafkax] producer and consumer both disabled")
	ErrClientClosed           = errors.New("[kafkax] client closed")
	ErrProducerClosed         = errors.New("[kafkax-producer] producer closed")
	ErrProducerNotInitialized = errors.New("[kafkax-producer] not initialized")
	ErrEmptyTopic             = errors.New("[kafkax-producer] empty topic")
	ErrNoTopics               = errors.New("[kafkax-consumer] no topics")
	ErrNoGroupID              = errors.New("[kafkax-consumer] no group id")
	ErrConsumerClosed         = errors.New("[kafkax-consumer] consumer closed")
	ErrConsumerNotInitialized = errors.New("[kafkax-consumer] not initialized")
)
