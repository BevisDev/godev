package rabbitmq

import (
	"time"

	"github.com/BevisDev/godev/utils/str"
)

type MsgProperties func(*msgProperties)

type msgProperties struct {
	messageID     string
	correlationID string
	persistentMsg bool
	expiration    string
	timestamp     time.Time
	headers       map[string]any
	userID        string
	appID         string
	replyTo       string
}

func WithMessageID(msgID string) MsgProperties {
	return func(p *msgProperties) {
		p.messageID = msgID
	}
}

func WithCorrelationID(correlationID string) MsgProperties {
	return func(p *msgProperties) {
		p.correlationID = correlationID
	}
}

func WithTimestamp(ts time.Time) MsgProperties {
	return func(p *msgProperties) {
		p.timestamp = ts
	}
}

func WithHeaders(headers map[string]any) MsgProperties {
	return func(p *msgProperties) {
		p.headers = headers
	}
}

func WithUserID(userID string) MsgProperties {
	return func(p *msgProperties) {
		p.userID = userID
	}
}

func WithPersistentMsg() MsgProperties {
	return func(o *msgProperties) {
		o.persistentMsg = true
	}
}

// WithExpiration sets message TTL. RabbitMQ expects expiration as milliseconds (string).
func WithExpiration(d time.Duration) MsgProperties {
	return func(o *msgProperties) {
		if d > 0 {
			o.expiration = str.ToString(d.Milliseconds())
		}
	}
}

func WithAppID(appID string) MsgProperties {
	return func(p *msgProperties) {
		p.appID = appID
	}
}

func WithReplyTo(replyTo string) MsgProperties {
	return func(p *msgProperties) {
		p.replyTo = replyTo
	}
}
