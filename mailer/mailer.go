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

	"github.com/BevisDev/godev/consts"
)

// Mailer handles email sending.
type Mailer struct {
	cfg  *Config
	auth smtp.Auth
	addr string
}

// Mail represents an email message.
type Mail struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []Attachment
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename string
	Content  []byte
}

// New creates a Mailer with the given config.
func New(cfg *Config) (*Mailer, error) {
	if cfg == nil {
		return nil, ErrConfigNil
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return &Mailer{
		cfg:  cfg,
		auth: auth,
		addr: addr,
	}, nil
}

// Send sends an email. It validates the mail and returns an error if send fails.
func (m *Mailer) Send(mail Mail) error {
	if err := validateMail(mail); err != nil {
		return err
	}

	message, err := m.buildMessage(mail)
	if err != nil {
		return err
	}

	return smtp.SendMail(
		m.addr,
		m.auth,
		m.cfg.From,
		append(mail.To, mail.Cc...),
		message,
	)
}

func validateMail(mail Mail) error {
	if len(mail.To) == 0 {
		return ErrNoRecipients
	}
	if mail.Subject == "" {
		return ErrEmptySubject
	}
	if mail.Body == "" {
		return ErrEmptyBody
	}
	return nil
}

// buildMessage constructs the full email message (headers + body, with optional attachments).
func (m *Mailer) buildMessage(mail Mail) ([]byte, error) {
	var buf bytes.Buffer

	m.writeHeaders(&buf, mail)

	if len(mail.Attachments) == 0 {
		m.writeSimpleBody(&buf, mail)
		return buf.Bytes(), nil
	}

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary)

	m.writeBodyPart(&buf, mail, boundary)
	for _, att := range mail.Attachments {
		m.writeAttachmentPart(&buf, att, boundary)
	}
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)

	return buf.Bytes(), nil
}

func (m *Mailer) writeHeaders(buf *bytes.Buffer, mail Mail) {
	fmt.Fprintf(buf, "From: %s\r\n", m.cfg.From)
	fmt.Fprintf(buf, "To: %s\r\n", strings.Join(mail.To, ", "))
	if len(mail.Cc) > 0 {
		fmt.Fprintf(buf, "Cc: %s\r\n", strings.Join(mail.Cc, ", "))
	}
	fmt.Fprintf(buf, "Subject: %s\r\n", mail.Subject)
	fmt.Fprintf(buf, "MIME-Version: 1.0\r\n")
}

func (m *Mailer) writeSimpleBody(buf *bytes.Buffer, mail Mail) {
	if mail.IsHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	buf.WriteString("\r\n")
	buf.WriteString(mail.Body)
}

func (m *Mailer) writeBodyPart(buf *bytes.Buffer, mail Mail, boundary string) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	if mail.IsHTML {
		fmt.Fprintf(buf, "Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		fmt.Fprintf(buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	}
	fmt.Fprintf(buf, "\r\n")
	fmt.Fprintf(buf, "%s\r\n", mail.Body)
}

func (m *Mailer) writeAttachmentPart(buf *bytes.Buffer, att Attachment, boundary string) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)

	contentType := mime.TypeByExtension(filepath.Ext(att.Filename))
	if contentType == "" {
		contentType = consts.ApplicationOctetStream
	}

	fmt.Fprintf(buf, "Content-Type: %s; name=\"%s\"\r\n", contentType, att.Filename)
	fmt.Fprintf(buf, "Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename)
	fmt.Fprintf(buf, "Content-Transfer-Encoding: base64\r\n\r\n")

	encoded := base64.StdEncoding.EncodeToString(att.Content)
	const lineLen = 76
	for i := 0; i < len(encoded); i += lineLen {
		end := min(i+lineLen, len(encoded))
		fmt.Fprintf(buf, "%s\r\n", encoded[i:end])
		fmt.Fprintf(buf, "\r\n")
	}
}

// SendTemplate sends an email by rendering the template at templatePath with data.
func (m *Mailer) SendTemplate(to []string, subject string, templatePath string, data any) error {
	body, err := executeTemplateFile(templatePath, data)
	if err != nil {
		return err
	}
	return m.Send(Mail{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// SendTemplateString sends an email by rendering the template string with data.
func (m *Mailer) SendTemplateString(to []string, subject string, templateStr string, data any) error {
	body, err := executeTemplateString(templateStr, data)
	if err != nil {
		return err
	}
	return m.Send(Mail{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

func executeTemplateFile(templatePath string, data any) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}
	return buf.String(), nil
}

func executeTemplateString(templateStr string, data any) (string, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateParse, err)
	}
	return buf.String(), nil
}
