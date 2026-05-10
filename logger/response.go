package logger

import "time"

type ResponseLogger struct {
	RID      string
	Duration time.Duration
	Status   int
	Header   any
	Body     string
}
