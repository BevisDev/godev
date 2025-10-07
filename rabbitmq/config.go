package rabbitmq

type Config struct {
	Host     string // RabbitMQ server host
	Port     int    // RabbitMQ server port
	Username string // Username for authentication
	Password string // Password for authentication
	VHost    string // VHost Virtual host
}
