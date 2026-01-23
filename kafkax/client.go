package kafkax

type Client interface {
	Close()
}

type Producer interface {
	Close()
}

type Consumer interface {
	Close()
}
