package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type Message struct {
	d amqp.Delivery
}

func (m Message) GetBody() []byte {
	return m.d.Body
}

func (m Message) Header(key string) any {
	return m.d.Headers[key]
}

func (m Message) Commit() {
	m.d.Ack(false)
}

func (m Message) CommitMulti() {
	m.d.Ack(true)
}

func (m Message) Requeue() {
	m.d.Nack(false, true)
}

func (m Message) RequeueMulti() {
	m.d.Nack(true, true)
}

func (m Message) Reject() {
	m.d.Reject(false)
}

func (m Message) RejectRequeue() {
	m.d.Reject(true)
}
