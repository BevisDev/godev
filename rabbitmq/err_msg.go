package rabbitmq

import "errors"

var (
	// rabbitmq
	ErrNilConfig         = errors.New("[rabbitmq] config is nil")
	ErrConnectionClosed  = errors.New("[rabbitmq]: connection is closed")
	ErrClientClosed      = errors.New("[rabbitmq]: client is already closed")
	ErrMaxRetriesReached = errors.New("[rabbitmq]: max connection retries reached")

	// queue
	ErrEmptyQueueName      = errors.New("[queue] name cannot be empty")
	ErrEmptyExchangeName   = errors.New("[queue] exchange name cannot be empty")
	ErrInvalidExchangeType = errors.New("[queue] invalid exchange type")
	ErrEmptyBindingQueue   = errors.New("[queue] binding queue name cannot be empty")

	// publisher
	ErrMessageTooLarge = errors.New("[publisher] message exceeds maximum size limit")
	ErrInvalidMessage  = errors.New("[publisher] invalid message format")
)
