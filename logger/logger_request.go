package logger

import "time"

type RequestLogger struct {
	State       string
	URL         string
	RequestTime time.Time
	Query       string
	Method      string
	Header      any
	Body        string
}
