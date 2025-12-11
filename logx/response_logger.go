package logx

import "time"

type ResponseLogger struct {
	State    string
	Duration time.Duration
	Status   int
	Header   any
	Body     string
}
