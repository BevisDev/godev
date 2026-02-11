package mailer

import (
	"errors"
	"strings"
	"testing"
)

func testConfig() *Config {
	return &Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "test@example.com",
		Password: "secret",
		From:     "noreply@example.com",
	}
}

func TestNewMailer(t *testing.T) {
	_, err := New(nil)
	if !errors.Is(err, ErrConfigNil) {
		t.Fatalf("expected ErrConfigNil, got %v", err)
	}

	m, err := New(testConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("mailer should not be nil")
	}
}

func TestSendValidation(t *testing.T) {
	m, _ := New(testConfig())

	tests := []struct {
		name string
		mail Mail
		err  error
	}{
		{
			"no recipients",
			Mail{Subject: "Hi", Body: "Body"},
			ErrNoRecipients,
		},
		{
			"empty subject",
			Mail{To: []string{"a@test.com"}, Body: "Body"},
			ErrEmptySubject,
		},
		{
			"empty body",
			Mail{To: []string{"a@test.com"}, Subject: "Hi"},
			ErrEmptyBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.Send(tt.mail)
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected %v, got %v", tt.err, err)
			}
		})
	}
}

func TestBuildMessage(t *testing.T) {
	m := &Mailer{
		cfg: &Config{
			From: "sender@example.com",
			Host: "smtp.example.com",
		},
	}

	t.Run("simple text email", func(t *testing.T) {
		mail := Mail{
			To:      []string{"recipient@example.com"},
			Subject: "Test",
			Body:    "Hello World",
			IsHTML:  false,
		}

		message, err := m.buildMessage(mail)
		if err != nil {
			t.Fatalf("buildMessage() error = %v", err)
		}

		msgStr := string(message)
		if !strings.Contains(msgStr, "Hello World") {
			t.Errorf("Message should contain body text")
		}
		if !strings.Contains(msgStr, "From:") {
			t.Errorf("Message should contain From header")
		}
		if !strings.Contains(msgStr, "To:") {
			t.Errorf("Message should contain To header")
		}
	})

	t.Run("HTML email", func(t *testing.T) {
		mail := Mail{
			To:      []string{"recipient@example.com"},
			Subject: "Test HTML",
			Body:    "<h1>Hello</h1>",
			IsHTML:  true,
		}

		message, err := m.buildMessage(mail)
		if err != nil {
			t.Fatalf("buildMessage() error = %v", err)
		}

		msgStr := string(message)
		if !strings.Contains(msgStr, "text/html") {
			t.Errorf("HTML email should have text/html content type")
		}
	})

	t.Run("email with attachment", func(t *testing.T) {
		mail := Mail{
			To:      []string{"recipient@example.com"},
			Subject: "Test Attachment",
			Body:    "See attachment",
			IsHTML:  false,
			Attachments: []Attachment{
				{
					Filename: "test.txt",
					Content:  []byte("Test file content"),
				},
			},
		}

		message, err := m.buildMessage(mail)
		if err != nil {
			t.Fatalf("buildMessage() error = %v", err)
		}

		msgStr := string(message)
		if !strings.Contains(msgStr, "multipart/mixed") {
			t.Errorf("Email with attachment should use multipart/mixed")
		}
		if !strings.Contains(msgStr, "test.txt") {
			t.Errorf("Message should contain attachment filename")
		}
	})
}

func TestSendTemplateString(t *testing.T) {
	m, _ := New(testConfig())

	err := m.SendTemplateString(
		[]string{"a@test.com"},
		"Hello",
		"Hello {{.Name}}",
		map[string]string{"Name": "Go"},
	)

	// Sẽ fail ở smtp.SendMail => nhưng không phải lỗi template
	if err == nil {
		t.Fatal("expected error from smtp.SendMail, got nil")
	}
	if strings.Contains(err.Error(), "template") {
		t.Fatalf("unexpected template error: %v", err)
	}
}

// Benchmark tests
func BenchmarkBuildMessage(b *testing.B) {
	m := &Mailer{
		cfg: &Config{
			From: "sender@example.com",
			Host: "smtp.example.com",
		},
	}

	mail := Mail{
		To:      []string{"recipient@example.com"},
		Subject: "Benchmark Test",
		Body:    "This is a benchmark test message",
		IsHTML:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.buildMessage(mail)
	}
}
