package ratelimit

type Mode int

const (
	AllowMode Mode = iota + 1
	WaitMode
)
