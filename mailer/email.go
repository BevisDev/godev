package mailer

// Email represents an email message.
type Email struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []*Attachment
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename string
	Content  []byte
}

func (e *Email) validate() error {
	if len(e.To) == 0 {
		return ErrNoRecipients
	}
	if e.Subject == "" {
		return ErrEmptySubject
	}
	if e.Body == "" {
		return ErrEmptyBody
	}
	return nil
}
