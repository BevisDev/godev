package redis

import "errors"

var (
	ErrMissingKey         = errors.New("use Key() before")
	ErrMissingKeys        = errors.New("use Keys() before")
	ErrMissingPrefix      = errors.New("use Prefix() before")
	ErrMissingValue       = errors.New("use Value() before")
	ErrMissingValues      = errors.New("use Values() before")
	ErrMissingChannel     = errors.New("use Channel() before")
	ErrMissingPushOrBatch = errors.New("use Push() or Batch() before")
)
