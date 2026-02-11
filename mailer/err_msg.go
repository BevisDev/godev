package mailer

import "errors"

// Errors
var (
	ErrConfigNil     = errors.New("[mailer] config is nil")
	ErrNoRecipients  = errors.New("[mailer] no recipients specified")
	ErrEmptySubject  = errors.New("[mailer] subject is empty")
	ErrEmptyBody     = errors.New("[mailer] body is empty")
	ErrTemplateParse = errors.New("[mailer] failed to parse template")
)
