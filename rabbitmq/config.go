package rabbitmq

import "fmt"

type Config struct {
	Host     string // RabbitMQ server host
	Port     int    // RabbitMQ server port
	Username string // Username for authentication
	Password string // Password for authentication
	VHost    string // VHost Virtual host
}

func (c *Config) clone() *Config {
	cc := *c
	return &cc
}

func (c *Config) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.VHost)
}
