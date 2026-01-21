package rabbitmq

type OptionFunc func(*options)

type options struct {
	persistentMsg bool // persistentMsg config to set delivery mode when publish message
}

func withDefaults() *options {
	return &options{
		persistentMsg: false,
	}
}

func WithPersistentMsg() OptionFunc {
	return func(o *options) {
		o.persistentMsg = true
	}
}
