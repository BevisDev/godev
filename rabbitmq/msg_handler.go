package rabbitmq

import (
	"fmt"
	"time"

	"github.com/BevisDev/godev/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MsgHandler struct {
	queueName string
	d         amqp.Delivery
	acked     bool
}

func NewMsgHandler(queueName string, d amqp.Delivery) *MsgHandler {
	return &MsgHandler{
		queueName: queueName,
		d:         d,
	}
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

// IsAcked reports whether ack/nack/reject has already been called.
func (m *MsgHandler) IsAcked() bool {
	return m.acked
}

// BodyAs decodes the message body (produced by utils.ToBytes) into type T.
func BodyAs[T any](m *MsgHandler) (T, error) {
	return utils.ValueFromBytes[T](m.d.Body)
}

func (m *MsgHandler) Header(key string) any {
	if m.d.Headers == nil {
		return nil
	}
	return m.d.Headers[key]
}

// HeaderString is a convenience accessor that returns the header as string.
// Returns ("", false) if the key is missing or not a string.
func (m *MsgHandler) HeaderString(key string) (string, bool) {
	v, ok := m.Header(key).(string)
	return v, ok
}

// GetHeaders returns a copy of all headers.
func (m *MsgHandler) GetHeaders() map[string]any {
	if m.d.Headers == nil {
		return nil
	}
	out := make(map[string]any, len(m.d.Headers))
	for k, v := range m.d.Headers {
		out[k] = v
	}
	return out
}

// Commit acks this message only.
func (m *MsgHandler) Commit() error {
	if m.acked {
		return fmt.Errorf(ErrMsgAcked, m.queueName)
	}
	if err := m.d.Ack(false); err != nil {
		return err
	}
	m.acked = true
	return nil
}

// CommitMulti acks this message AND all previous unacked messages on the same channel
// Use only when you know all earlier messages are safe to ack.
// Typically called on the LAST message of a batch.
func (m *MsgHandler) CommitMulti() error {
	if m.acked {
		return fmt.Errorf(ErrMsgAcked, m.queueName)
	}
	if err := m.d.Ack(true); err != nil {
		return err
	}
	m.acked = true
	return nil
}

// Requeue nacks this message and asks the broker to requeue it.
func (m *MsgHandler) Requeue() error {
	if m.acked {
		return fmt.Errorf(ErrMsgAcked, m.queueName)
	}
	if err := m.d.Nack(false, true); err != nil {
		return err
	}
	m.acked = true
	return nil
}

// RequeueMulti nacks this message + all previous unacked, with requeue=true.
func (m *MsgHandler) RequeueMulti() error {
	if m.acked {
		return fmt.Errorf(ErrMsgAcked, m.queueName)
	}
	if err := m.d.Nack(true, true); err != nil {
		return err
	}
	m.acked = true
	return nil
}

// Reject discards this message (will be dead-lettered if DLX configured).
func (m *MsgHandler) Reject() error {
	if m.acked {
		return fmt.Errorf(ErrMsgAcked, m.queueName)
	}
	if err := m.d.Reject(false); err != nil {
		return err
	}
	m.acked = true
	return nil
}

// RejectRequeue rejects this message and asks the broker to requeue it.
// Note: prefer Requeue() for clarity.
func (m *MsgHandler) RejectRequeue() error {
	if m.acked {
		return fmt.Errorf(ErrMsgAcked, m.queueName)
	}
	if err := m.d.Reject(true); err != nil {
		return err
	}
	m.acked = true
	return nil
}
