package kafkax

// About configuration
// https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
// * = both producer and consumer
const (
	clientId              = "client.id"
	bootstrapServers      = "bootstrap.servers"
	reconnectBackoffMs    = "reconnect.backoff.ms"
	reconnectBackoffMaxMs = "reconnect.backoff.max.ms"
	connectionMaxIdleMs   = "connections.max.idle.ms"
)

// producer
const (
	retries = "retries"
	acks    = "acks"

	// if enableIdempotence = true
	// the configuration are adjusted automatically (if not modified by the user)
	// max.in.flight.requests.per.connection=5 (must <=5)
	// retries=INT32_MAX (must >0)
	// acks=all
	// queuing.strategy=fifo
	enableIdempotence = "enable.idempotence"
	messageMaxBytes   = "message.max.bytes"
	requestTimeoutMs  = "request.timeout.ms"
	deliveryTimeoutMs = "delivery.timeout.ms"
)

// consumer
const (
	groupId          = "group.id"
	autoOffsetReset  = "auto.offset.reset"
	enableAutoCommit = "enable.auto.commit"
	sessionTimeoutMs = "session.timeout.ms"
)

type AutoOffsetReset string

const (
	Earliest  = AutoOffsetReset("earliest")
	Beginning = AutoOffsetReset("beginning")
	Latest    = AutoOffsetReset("latest")
	Error     = AutoOffsetReset("error")
)

type Config struct {
	ClientId         string
	BootstrapServers string
	ProducerConfig   *ProducerConfig
	ConsumerConfig   *ConsumerConfig
}

type ProducerConfig struct {
	Retries           int
	Acks              int
	EnableIdempotence bool
	MessageMaxBytes   int
	RequestTimeoutMs  int
	DeliveryTimeoutMs int
}

type ConsumerConfig struct {
	GroupID          string
	AutoOffsetReset  AutoOffsetReset
	EnableAutoCommit bool
}
