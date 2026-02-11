# Mailer Package

A robust, production-ready email sending package for Go with support for HTML templates, attachments, TLS/SSL, and batch sending.

## Features

✅ **Simple API** - Easy to use, clean interface  
✅ **HTML & Text Emails** - Support for both formats  
✅ **Templates** - File, string, and embedded FS template support  
✅ **Attachments** - Multiple attachments with inline images  
✅ **TLS/SSL** - Secure connections with configurable TLS  
✅ **Batch Sending** - Concurrent batch email sending  
✅ **Priority Levels** - High, normal, low priority emails  
✅ **Custom Headers** - Add custom email headers  
✅ **Validation** - Comprehensive input validation  
✅ **Error Handling** - Detailed error messages  
✅ **Well Tested** - Extensive unit tests

## Installation

```bash
go get github.com/yourusername/mailer
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/yourusername/mailer"
)

func main() {
    // Configure mailer
    cfg := &mailer.Config{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        From:     "noreply@example.com",
        FromName: "My App",
        UseTLS:   true,
    }

    // Create mailer instance
    m, err := mailer.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Send email
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

### Basic Configuration

```go
cfg := &mailer.Config{
    Host:     "smtp.gmail.com",  // SMTP server
    Port:     587,                // SMTP port (587 for TLS, 465 for SSL)
    Username: "user@gmail.com",   // SMTP username
    Password: "app-password",     // SMTP password
    From:     "noreply@example.com", // From address
}
```

### Advanced Configuration

```go
cfg := &mailer.Config{
    Host:               "smtp.gmail.com",
    Port:               587,
    Username:           "user@gmail.com",
    Password:           "app-password",
    From:               "noreply@example.com",
    FromName:           "My Application",        // Display name
    UseTLS:             true,                     // Use TLS/SSL
    InsecureSkipVerify: false,                   // Skip SSL verification (dev only)
    Timeout:            30 * time.Second,        // Connection timeout
    KeepAlive:          30 * time.Second,        // Keep alive duration
}
```

### Gmail Configuration

For Gmail, you need to:
1. Enable 2-factor authentication
2. Generate an App Password
3. Use the App Password instead of your regular password

```go
cfg := &mailer.Config{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "your-email@gmail.com",
    Password: "your-16-char-app-password", // From Google Account settings
    From:     "your-email@gmail.com",
    UseTLS:   true,
}
```

## Usage Examples

### 1. Simple Text Email

```go
mail := mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Welcome!",
    Body:    "Thank you for signing up.",
    IsHTML:  false,
}

err := m.Send(mail)
```

### 2. HTML Email

```go
htmlBody := `
<!DOCTYPE html>
<html>
<body>
    <h1>Welcome!</h1>
    <p>Thank you for joining us.</p>
    <a href="https://example.com">Get Started</a>
</body>
</html>
`

mail := mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Welcome to Our Service",
    Body:    htmlBody,
    IsHTML:  true,
}

err := m.Send(mail)
```

### 3. Email with Multiple Recipients

```go
mail := mailer.Mail{
    To:      []string{"user1@example.com", "user2@example.com"},
    Cc:      []string{"manager@example.com"},
    Bcc:     []string{"admin@example.com"},
    Subject: "Team Update",
    Body:    "<h1>Important Update</h1>",
    IsHTML:  true,
}

err := m.Send(mail)
```

### 4. Email with Attachments

```go
// Read file
pdfContent, _ := os.ReadFile("document.pdf")

mail := mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Your Document",
    Body:    "<p>Please find attached document.</p>",
    IsHTML:  true,
    Attachments: []mailer.Attachment{
        {
            Filename:    "document.pdf",
            Content:     pdfContent,
            ContentType: "application/pdf",
        },
        {
            Filename:    "data.csv",
            Content:     []byte("Name,Email\nJohn,john@example.com"),
            ContentType: "text/csv",
        },
    },
}

err := m.Send(mail)
```

### 5. Email with Inline Images

```go
logoContent, _ := os.ReadFile("logo.png")

htmlBody := `
<html>
<body>
    <img src="cid:company-logo" alt="Logo">
    <h1>Welcome!</h1>
</body>
</html>
`

mail := mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Welcome",
    Body:    htmlBody,
    IsHTML:  true,
    Attachments: []mailer.Attachment{
        {
            Filename:    "logo.png",
            Content:     logoContent,
            ContentType: "image/png",
            Inline:      true,
            ContentID:   "company-logo",
        },
    },
}

err := m.Send(mail)
```

### 6. High Priority Email

```go
mail := mailer.Mail{
    To:       []string{"admin@example.com"},
    Subject:  "URGENT: System Alert",
    Body:     "<h1>Critical Error</h1>",
    IsHTML:   true,
    Priority: mailer.PriorityHigh,
}

err := m.Send(mail)
```

### 7. Email with Custom Headers

```go
mail := mailer.Mail{
    To:      []string{"user@example.com"},
    Subject: "Custom Email",
    Body:    "Email with custom headers",
    Headers: map[string]string{
        "X-Custom-ID":    "12345",
        "X-Campaign":     "newsletter-2024",
        "X-Tracking-ID":  "abc123",
    },
    ReplyTo: "support@example.com",
}

err := m.Send(mail)
```

### 8. Template Email (File)

Create template file `templates/welcome.html`:
```html
<!DOCTYPE html>
<html>
<body>
    <h1>Hello {{.Name}}!</h1>
    <p>Your account ID is: {{.AccountID}}</p>
    <p>Click <a href="{{.VerifyURL}}">here</a> to verify.</p>
</body>
</html>
```

Send email:
```go
data := struct {
    Name      string
    AccountID string
    VerifyURL string
}{
    Name:      "John Doe",
    AccountID: "ACC123",
    VerifyURL: "https://example.com/verify?token=xyz",
}

err := m.SendTemplate(
    []string{"user@example.com"},
    "Welcome to Our Service",
    "templates/welcome.html",
    data,
)
```

### 9. Template Email (String)

```go
templateStr := `
<html>
<body>
    <h1>Hi {{.Username}}!</h1>
    <p>Your verification code: <strong>{{.Code}}</strong></p>
    <p>Expires in {{.Minutes}} minutes</p>
</body>
</html>
`

data := struct {
    Username string
    Code     string
    Minutes  int
}{
    Username: "johndoe",
    Code:     "123456",
    Minutes:  15,
}

err := m.SendTemplateString(
    []string{"user@example.com"},
    "Verification Code",
    templateStr,
    data,
)
```

### 10. Template Email (Embedded FS)

```go
//go:embed templates/*
var templatesFS embed.FS

data := struct {
    OrderID string
    Total   float64
}{
    OrderID: "ORD-12345",
    Total:   299.99,
}

err := m.SendTemplateFS(
    []string{"user@example.com"},
    "Order Confirmation",
    templatesFS,
    "templates/order_confirmation.html",
    data,
)
```

### 11. Batch Sending

```go
// Prepare emails
mails := []mailer.Mail{
    {
        To:      []string{"user1@example.com"},
        Subject: "Newsletter",
        Body:    "<h1>News 1</h1>",
        IsHTML:  true,
    },
    {
        To:      []string{"user2@example.com"},
        Subject: "Newsletter",
        Body:    "<h1>News 2</h1>",
        IsHTML:  true,
    },
    // ... more emails
}

// Send with 5 concurrent workers
errors := m.SendBatch(mails, 5)

// Check results
for i, err := range errors {
    if err != nil {
        log.Printf("Failed to send email %d: %v", i, err)
    }
}
```

## Error Handling

The package provides specific error types:

```go
err := m.Send(mail)
if err != nil {
    switch {
    case errors.Is(err, mailer.ErrNoRecipients):
        log.Println("No recipients specified")
    case errors.Is(err, mailer.ErrInvalidEmail):
        log.Println("Invalid email address")
    case errors.Is(err, mailer.ErrEmptySubject):
        log.Println("Subject is required")
    case errors.Is(err, mailer.ErrEmptyBody):
        log.Println("Body is required")
    case errors.Is(err, mailer.ErrSendFailed):
        log.Println("Failed to send email")
    default:
        log.Printf("Error: %v", err)
    }
}
```

## Best Practices

### 1. Use Environment Variables for Credentials

```go
cfg := &mailer.Config{
    Host:     os.Getenv("SMTP_HOST"),
    Port:     getEnvInt("SMTP_PORT", 587),
    Username: os.Getenv("SMTP_USERNAME"),
    Password: os.Getenv("SMTP_PASSWORD"),
    From:     os.Getenv("SMTP_FROM"),
    UseTLS:   true,
}
```

### 2. Implement Retry Logic

```go
func sendWithRetry(m *mailer.Mailer, mail mailer.Mail, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = m.Send(mail)
        if err == nil {
            return nil
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
    return err
}
```

### 3. Use Connection Pooling for Batch Sending

```go
// Already implemented in SendBatch method
errors := m.SendBatch(mails, 10) // 10 concurrent workers
```

### 4. Validate Emails Before Sending

```go
// The package validates automatically, but you can pre-validate:
if !isValidEmail(email) {
    return errors.New("invalid email")
}
```

### 5. Use Templates for Consistent Branding

```go
// Store templates in embedded FS for consistency
//go:embed templates/*
var templates embed.FS

// Reuse templates across your application
```

## Testing

Run tests:
```bash
go test -v
```

Run tests with coverage:
```bash
go test -cover
```

Run benchmarks:
```bash
go test -bench=.
```

## Common SMTP Providers

### Gmail
```go
Host: "smtp.gmail.com"
Port: 587 (TLS) or 465 (SSL)
```

### Outlook/Office365
```go
Host: "smtp.office365.com"
Port: 587
```

### Yahoo
```go
Host: "smtp.mail.yahoo.com"
Port: 587
```

### SendGrid
```go
Host: "smtp.sendgrid.net"
Port: 587
Username: "apikey"
Password: "your-api-key"
```

### Mailgun
```go
Host: "smtp.mailgun.org"
Port: 587
Username: "postmaster@your-domain.com"
Password: "your-smtp-password"
```

## Troubleshooting

### Gmail "Less secure app access" Error
- Enable 2FA on your Google Account
- Generate an App Password
- Use the App Password instead of your regular password

### Connection Timeout
- Increase the timeout in config
- Check firewall settings
- Verify SMTP server is accessible

### Authentication Failed
- Verify username and password
- Check if 2FA is enabled (use app password)
- Ensure SMTP access is enabled

### TLS/SSL Errors
- Use correct port (587 for TLS, 465 for SSL)
- Set `UseTLS: true` for port 587
- For development, you can use `InsecureSkipVerify: true`

## License

MIT License - feel free to use in your projects!

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you encounter any issues or have questions, please open an issue on GitHub.