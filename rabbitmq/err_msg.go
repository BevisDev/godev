package rabbitmq

import "errors"

var (
	ErrNilConfig         = errors.New("[rabbitmq] config is nil")
	ErrConnectionClosed  = errors.New("[rabbitmq]: connection is closed")
	ErrClientClosed      = errors.New("[rabbitmq]: client is already closed")
	ErrMaxRetriesReached = errors.New("[rabbitmq]: max connection retries reached")
)
