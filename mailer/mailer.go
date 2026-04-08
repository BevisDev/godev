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
	"github.com/BevisDev/godev/utils/crypto"
	"github.com/BevisDev/godev/utils/datetime"
)

type Mailer interface {
	Send(mail *Email) error
	SendTemplate(to []string, subject string, templatePath string, data any) error
	SendTemplateString(to []string, subject string, templateStr string, data any) error
}

// smtpMailer handles email sending.
type smtpMailer struct {
	cfg  *Config
	auth smtp.Auth
	addr string
}

// New creates a Mailer with the given config.
func New(cfg *Config) (*smtpMailer, error) {
	if cfg == nil {
		return nil, ErrConfigNil
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return &smtpMailer{
		cfg:  cfg,
		auth: auth,
		addr: addr,
	}, nil
}

// Send sends an email. It validates the mail and returns an error if send fails.
func (m *smtpMailer) Send(mail *Email) error {
	if err := mail.validate(); err != nil {
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
		mail.To,
		message,
	)
}

// buildMessage constructs the full email message (headers + body, with optional attachments).
func (m *smtpMailer) buildMessage(mail *Email) ([]byte, error) {
	var buf bytes.Buffer
	var boundary string
	if len(mail.Attachments) > 0 {
		boundary = fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	}

	// write headers
	m.writeHeaders(&buf, mail, boundary)

	// write body
	if boundary == "" {
		buf.WriteString(mail.Body + consts.CRLF)
	} else {
		m.writeBodyPart(&buf, mail, boundary)
		for _, att := range mail.Attachments {
			m.writeAttachmentPart(&buf, att, boundary)
		}
		fmt.Fprintf(&buf, "--%s--%s", boundary, consts.CRLF)
	}

	return buf.Bytes(), nil
}

func (m *smtpMailer) writeHeaders(buf *bytes.Buffer, mail *Email, boundary string) {
	fmt.Fprintf(buf, "%s: %s%s", consts.From, m.cfg.From, consts.CRLF)
	fmt.Fprintf(buf, "%s: %s%s",
		consts.To, strings.Join(mail.To, ", "), consts.CRLF)
	if len(mail.Cc) > 0 {
		fmt.Fprintf(buf, "%s: %s%s",
			consts.Cc, strings.Join(mail.Cc, ", "), consts.CRLF)
	}

	subjectEncoded := base64.StdEncoding.EncodeToString([]byte(mail.Subject))
	fmt.Fprintf(buf, "%s: =?UTF-8?B?%s?=%s",
		consts.Subject, subjectEncoded, consts.CRLF)

	fmt.Fprintf(buf, "MIME-Version: 1.0%s", consts.CRLF)
	fmt.Fprintf(buf, "%s: %s%s",
		consts.Date, datetime.ToString(time.Now(), time.RFC1123Z), consts.CRLF)

	if boundary != "" {
		fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=\"%s\"%s", boundary, consts.CRLF)
	} else {
		contentType := consts.TextPlain
		if mail.IsHTML {
			contentType = consts.TextHTML
		}
		fmt.Fprintf(buf, "%s: %s; %s%s",
			consts.ContentType, contentType, consts.CharsetUTF8, consts.CRLF)
	}
	buf.WriteString(consts.CRLF)
}

func (m *smtpMailer) writeBodyPart(buf *bytes.Buffer, mail *Email, boundary string) {
	fmt.Fprintf(buf, "--%s%s", boundary, consts.CRLF)
	contentType := consts.TextPlain
	if mail.IsHTML {
		contentType = consts.TextHTML
	}
	fmt.Fprintf(buf, "%s: %s; %s%s",
		consts.ContentType, contentType, consts.CharsetUTF8, consts.CRLF)
	buf.WriteString(consts.CRLF)
	buf.WriteString(mail.Body + consts.CRLF)
}

func (m *smtpMailer) writeAttachmentPart(buf *bytes.Buffer, att *Attachment, boundary string) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)

	contentType := mime.TypeByExtension(filepath.Ext(att.Filename))
	if contentType == "" {
		contentType = consts.ApplicationOctetStream
	}

	fmt.Fprintf(buf, `%s: %s; name="%s"%s`,
		consts.ContentType, contentType, att.Filename, consts.CRLF)
	fmt.Fprintf(buf, `%s: %s%s`,
		consts.ContentDisposition, fmt.Sprintf(consts.ContentDispositionAttachment, att.Filename), consts.CRLF)
	fmt.Fprintf(buf, "%s: base64%s",
		consts.ContentTransferEncoding, consts.CRLF)
	buf.WriteString(consts.CRLF)

	encoded := crypto.EncodeBase64Bytes(att.Content)
	const lineLen = 76
	for i := 0; i < len(encoded); i += lineLen {
		end := min(i+lineLen, len(encoded))
		fmt.Fprintf(buf, "%s\r\n", encoded[i:end])
		fmt.Fprintf(buf, "\r\n")
	}
}

// SendTemplate sends an email by rendering the template at templatePath with data.
func (m *smtpMailer) SendTemplate(to []string, subject string, templatePath string, data any) error {
	body, err := executeTemplateFile(templatePath, data)
	if err != nil {
		return err
	}
	return m.Send(&Email{
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// SendTemplateString sends an email by rendering the template string with data.
func (m *smtpMailer) SendTemplateString(to []string, subject string, templateStr string, data any) error {
	body, err := executeTemplateString(templateStr, data)
	if err != nil {
		return err
	}
	return m.Send(&Email{
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
