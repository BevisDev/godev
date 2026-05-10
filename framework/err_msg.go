package framework

import "errors"

var (
	ErrAlreadyInitialized = errors.New(prefixBootstrap + "already initialized")
	ErrAlreadyStarted     = errors.New(prefixBootstrap + "already started")
)
