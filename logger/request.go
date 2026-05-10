package logger

import "time"

type RequestLogger struct {
	RID    string
	URL    string
	Time   time.Time
	Query  string
	Method string
	Header any
	Body   string
}
