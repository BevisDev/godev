package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type MsgHandler struct {
	amqp.Delivery
}

func (m MsgHandler) GetBody() []byte {
	return m.Body
}

func (m MsgHandler) Header(key string) any {
	return m.Headers[key]
}

func (m MsgHandler) Commit() {
	m.Ack(false)
}

func (m MsgHandler) CommitMulti() {
	m.Ack(true)
}

func (m MsgHandler) Requeue() {
	m.Nack(false, true)
}

func (m MsgHandler) RequeueMulti() {
	m.Nack(true, true)
}
