package mailer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"
)

// Mailer handles email sending
type Mailer struct {
	cfg  *Config
	auth smtp.Auth
}

// Mail represents an email message
type Mail struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Content  []byte
}

func New(cfg *Config) (*Mailer, error) {
	if cfg == nil {
		return nil, ErrConfigNil
	}

	auth := smtp.PlainAuth(
		"",
		cfg.Username,
		cfg.Password,
		cfg.Host,
	)

	return &Mailer{
		cfg:  cfg,
		auth: auth,
	}, nil
}

func (m *Mailer) Send(mail Mail) error {
	if len(mail.To) == 0 {
		return ErrNoRecipients
	}
	if mail.Subject == "" {
		return ErrEmptySubject
	}
	if mail.Body == "" {
		return ErrEmptyBody
	}

	// Build message
	message, err := m.buildMessage(mail)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	return smtp.SendMail(
		addr,
		m.auth,
		m.cfg.From,
		append(mail.To, mail.Cc...),
		message,
	)
}

// buildMessage constructs the email message with all headers and content
func (m *Mailer) buildMessage(mail Mail) ([]byte, error) {
	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", m.cfg.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ", ")))

	if len(mail.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(mail.Cc, ", ")))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mail.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// No attachments - simple message
	if len(mail.Attachments) == 0 {
		if mail.IsHTML {
			buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		} else {
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		}
		buf.WriteString("\r\n")
		buf.WriteString(mail.Body)
		return buf.Bytes(), nil
	}

	// With attachments
	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

	// Body part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	if mail.IsHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	buf.WriteString("\r\n")
	buf.WriteString(mail.Body)
	buf.WriteString("\r\n")

	// Attachments
	for _, att := range mail.Attachments {
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))

		// Content type
		contentType := mime.TypeByExtension(filepath.Ext(att.Filename))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", contentType, att.Filename))
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

		// Encode base64
		encoded := base64.StdEncoding.EncodeToString(att.Content)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			buf.WriteString(encoded[i:end])
			buf.WriteString("\r\n")
		}
	}

	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return buf.Bytes(), nil
}

func (m *Mailer) buildHeaders(mail Mail) map[string]string {
	headers := map[string]string{
		"From":    m.cfg.From,
		"To":      strings.Join(mail.To, ";"),
		"Subject": mail.Subject,
	}

	// Cc
	if len(mail.Cc) > 0 {
		headers["Cc"] = strings.Join(mail.Cc, ", ")
	}

	// Content-Type (if no attachments)
	if len(mail.Attachments) == 0 {
		if mail.IsHTML {
			headers["Content-Type"] = "text/html; charset=UTF-8"
		} else {
			headers["Content-Type"] = "text/plain; charset=UTF-8"
		}
		headers["Content-Transfer-Encoding"] = "quoted-printable"
	}

	return headers
}

// SendTemplate sends an email using a template
func (m *Mailer) SendTemplate(
	to []string,
	subject string,
	templatePath string,
	data any,
) error {
	// Parse template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}

	// Send email
	mail := Mail{
		To:      to,
		Subject: subject,
		Body:    body.String(),
		IsHTML:  true,
	}

	return m.Send(mail)
}

// SendTemplateString sends email using template string
func (m *Mailer) SendTemplateString(to []string, subject string, templateStr string, data any) error {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}

	return m.Send(Mail{
		To:      to,
		Subject: subject,
		Body:    body.String(),
		IsHTML:  true,
	})
}
