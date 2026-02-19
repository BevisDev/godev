# Mailer Package

A small, focused email sending package for Go with SMTP, HTML/text support, templates, and attachments.

## Features

- **Simple API** – `New(cfg)`, `Send(mail)`, `SendTemplate`, `SendTemplateString`
- **HTML & plain text** – Set `IsHTML: true` for HTML body
- **Templates** – File-based or string-based templates with `html/template`
- **Attachments** – Multiple attachments with MIME type from file extension
- **Validation** – Clear errors: no recipients, empty subject, empty body, template parse failure

## Installation

```bash
go get github.com/BevisDev/godev/mailer
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/BevisDev/godev/mailer"
)

func main() {
    cfg := &mailer.Config{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        From:     "noreply@example.com",
    }

    m, err := mailer.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    err = m.Send(mailer.Mail{
        To:      []string{"recipient@example.com"},
        Subject: "Hello!",
        Body:    "This is a test email",
        IsHTML:  false,
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

```go
cfg := &mailer.Config{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "user@gmail.com",
    Password: "app-password",
    From:     "noreply@example.com",
}
```

For Gmail: enable 2FA and use an [App Password](https://support.google.com/accounts/answer/185833).

## Usage

### Plain text email

```go
err := m.Send(mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Welcome",
    Body:    "Thank you for signing up.",
    IsHTML:  false,
})
```

### HTML email

```go
err := m.Send(mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Welcome",
    Body:    "<h1>Hello</h1><p>Thanks for joining.</p>",
    IsHTML:  true,
})
```

### Multiple recipients (To, Cc)

```go
err := m.Send(mailer.Mail{
    To:      []string{"user1@example.com", "user2@example.com"},
    Cc:      []string{"manager@example.com"},
    Subject: "Update",
    Body:    "<h1>Update</h1>",
    IsHTML:  true,
})
```

### Attachments

```go
content, _ := os.ReadFile("document.pdf")

err := m.Send(mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Your document",
    Body:    "Please find the attachment.",
    IsHTML:  false,
    Attachments: []mailer.Attachment{
        {Filename: "document.pdf", Content: content},
    },
})
```

### Template from file

Template `templates/welcome.html`:

```html
<!DOCTYPE html>
<html>
<body>
    <h1>Hello {{.Name}}!</h1>
    <p>Verify: <a href="{{.VerifyURL}}">click here</a></p>
</body>
</html>
```

```go
data := struct {
    Name      string
    VerifyURL string
}{Name: "John", VerifyURL: "https://example.com/verify?token=xyz"}

err := m.SendTemplate(
    []string{"user@example.com"},
    "Welcome",
    "templates/welcome.html",
    data,
)
```

### Template from string

```go
err := m.SendTemplateString(
    []string{"user@example.com"},
    "Verification",
    "Hi {{.Name}}, your code is: <strong>{{.Code}}</strong>",
    map[string]string{"Name": "John", "Code": "123456"},
)
```

## Error handling

Use sentinel errors with `errors.Is`:

```go
err := m.Send(mail)
if err != nil {
    switch {
    case errors.Is(err, mailer.ErrConfigNil):
        // New(nil)
    case errors.Is(err, mailer.ErrNoRecipients):
        // no To addresses
    case errors.Is(err, mailer.ErrEmptySubject):
        // subject empty
    case errors.Is(err, mailer.ErrEmptyBody):
        // body empty
    case errors.Is(err, mailer.ErrTemplateParse):
        // SendTemplate / SendTemplateString template error
    default:
        // e.g. SMTP send failure
    }
}
```

## Testing

```bash
go test -v
go test -cover
go test -bench=.
```

Note: `TestSendTemplateString` performs a real SMTP call and will fail without a valid SMTP server (e.g. in CI). Run validation and build tests only: `go test -run 'TestNewMailer|TestSendValidation|TestBuildMessage'`.

## Common SMTP settings

| Provider   | Host                   | Port |
|-----------|------------------------|------|
| Gmail     | smtp.gmail.com         | 587  |
| Outlook   | smtp.office365.com     | 587  |
| SendGrid  | smtp.sendgrid.net      | 587  |
| Mailgun   | smtp.mailgun.org       | 587  |
