package database

import "errors"

var (
	ErrMissingFrom   = errors.New("use From() before")
	ErrMissingSelect = errors.New("use Select() before")
	ErrMissingWhere  = errors.New("use Where() before")
	ErrMissingTable  = errors.New("missing TableName() for model")
	ErrMissingData   = errors.New("missing model data")
)
