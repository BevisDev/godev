package datetime

type Unit int

const (
	Nanosecond Unit = iota + 1
	Millisecond
	Second
	Minute
	Hour
	Day
	Month
	Year
)
