package rabbitmq

type Option func(*options)

type options struct {
	persistentMsg bool // persistentMsg config to set delivery mode when publish message
}

func withDefaults() *options {
	return &options{
		persistentMsg: false,
	}
}

func WithPersistentMsg() Option {
	return func(o *options) {
		o.persistentMsg = true
	}
}
