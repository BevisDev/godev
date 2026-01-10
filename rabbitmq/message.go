package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type Message struct {
	amqp.Delivery
}

func (m Message) GetBody() []byte {
	return m.Body
}

func (m Message) Header(key string) any {
	return m.Headers[key]
}

func (m Message) Commit() {
	m.Ack(false)
}

func (m Message) CommitMulti() {
	m.Ack(true)
}

func (m Message) Requeue() {
	m.Nack(false, true)
}

func (m Message) RequeueMulti() {
	m.Nack(true, true)
}
