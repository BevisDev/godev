package rabbitmq

import (
	"time"

	"github.com/BevisDev/godev/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MsgHandler struct {
	queueName string
	d         amqp.Delivery
}

func (m *MsgHandler) QueueName() string {
	return m.queueName
}

func (m *MsgHandler) Timestamp() time.Time {
	return m.d.Timestamp
}

func (m *MsgHandler) ContentType() string {
	return m.d.ContentType
}

func (m *MsgHandler) CorrelationID() string {
	return m.d.CorrelationId
}

func (m *MsgHandler) GetBody() []byte {
	return m.d.Body
}

// BodyAs decodes the message body (produced by utils.ToBytes) into type T.
func BodyAs[T any](m *MsgHandler) (T, error) {
	return utils.ToValue[T](m.d.Body)
}

func (m *MsgHandler) Header(key string) any {
	return m.d.Headers[key]
}

func (m *MsgHandler) Commit() {
	m.d.Ack(false)
}

func (m *MsgHandler) CommitMulti() {
	m.d.Ack(true)
}

func (m *MsgHandler) Requeue() {
	m.d.Nack(false, true)
}

func (m *MsgHandler) RequeueMulti() {
	m.d.Nack(true, true)
}

func (m *MsgHandler) Reject() {
	m.d.Reject(false)
}

func (m *MsgHandler) RejectRequeue() {
	m.d.Reject(true)
}
