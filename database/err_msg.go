package database

import "errors"

var (
	ErrMissingFrom   = errors.New("use From() before")
	ErrMissingSelect = errors.New("use Select() before")
	ErrMissingWhere  = errors.New("use Where() before")
)
