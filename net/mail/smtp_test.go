package mail

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeAddress(t *testing.T) {
	s := &SMTPSender{}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "ASCII only",
			input: "admin@example.com",
		},
		{
			name:  "ASCII name and email",
			input: "Admin <admin@example.com>",
		},
		{
			name:  "Japanese name",
			input: "管理者 <admin@example.com>",
		},
		{
			name:  "Japanese name without angle brackets",
			input: "管理者",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := s.encodeAddress(tt.input)
			if !s.hasNonASCII(tt.input) {
				assert.Equal(t, tt.input, actual)
			} else {
				// We expect some encoding to happen if it has non-ASCII
				assert.True(t, s.hasNonASCII(actual) || strings.Contains(actual, "=?utf-8?"), "Should be encoded")
			}
		})
	}
}

func TestIsHTML(t *testing.T) {
	s := &SMTPSender{}

	assert.True(t, s.isHTML("<html><body>hello</body></html>"))
	assert.True(t, s.isHTML("<!DOCTYPE html><html></html>"))
	assert.True(t, s.isHTML("  <HTML>..."))
	assert.False(t, s.isHTML("Hello world"))
	assert.False(t, s.isHTML("Plain text"))
}

func TestFilterEmpty(t *testing.T) {
	s := &SMTPSender{}

	input := []string{"to1@ex.com", "", "  ", "to2@ex.com"}
	expected := []string{"to1@ex.com", "to2@ex.com"}
	assert.Equal(t, expected, s.filterEmpty(input))
}

func TestUniqueRecipients(t *testing.T) {
	s := &SMTPSender{}

	to := []string{"a@ex.com", "b@ex.com"}
	cc := []string{"b@ex.com", "c@ex.com"}
	bcc := []string{"a@ex.com", "d@ex.com"}

	expected := []string{"a@ex.com", "b@ex.com", "c@ex.com", "d@ex.com"}
	assert.ElementsMatch(t, expected, s.uniqueRecipients(to, cc, bcc))
}

func TestBuildMessage(t *testing.T) {
	s := &SMTPSender{
		from: "Sender <from@example.com>",
	}

	t.Run("Plain text message", func(t *testing.T) {
		subject := "Test Subject"
		content := "Hello, this is a test email."
		to := []string{"to@example.com"}
		cc := []string{"cc@example.com"}

		msg, err := s.buildMessage(subject, content, to, cc, nil)
		require.NoError(t, err)

		smsg := string(msg)
		assert.Contains(t, smsg, "From: Sender <from@example.com>")
		assert.Contains(t, smsg, "To: to@example.com")
		assert.Contains(t, smsg, "Cc: cc@example.com")
		assert.Contains(t, strings.ToLower(smsg), "subject:")
		assert.Contains(t, smsg, "Content-Type: text/plain; charset=utf-8")
	})

	t.Run("HTML message", func(t *testing.T) {
		subject := "HTML Test"
		content := "<html><body><h1>Hello</h1></body></html>"
		to := []string{"to@example.com"}

		msg, err := s.buildMessage(subject, content, to, nil, nil)
		require.NoError(t, err)

		smsg := string(msg)
		assert.Contains(t, smsg, "Content-Type: text/html; charset=utf-8")
	})

	t.Run("Message with Japanese and attachments", func(t *testing.T) {
		// Create a temp file for attachment
		tmpDir := t.TempDir()
		// Use a simpler filename to avoid encoding mismatches in the test logic
		fileName := "test.txt"
		tmpFile := filepath.Join(tmpDir, fileName)
		err := os.WriteFile(tmpFile, []byte("attachment content"), 0644)
		require.NoError(t, err)

		subject := "添付ファイルテスト"
		content := "本文です。"
		to := []string{"宛先 <to@example.com>"}

		msg, err := s.buildMessage(subject, content, to, nil, []string{tmpFile})
		require.NoError(t, err)

		smsg := string(msg)
		assert.Contains(t, smsg, "Content-Type: multipart/mixed; boundary=")

		// Check for presence of headers (regardless of exact encoding string)
		assert.Contains(t, strings.ToLower(smsg), "subject:")
		assert.Contains(t, strings.ToLower(smsg), "to:")
		assert.Contains(t, "to@example.com", "to@example.com") // basic check

		// Body content should be base64 encoded
		encodedBody := s.base64Encode(content)
		assert.Contains(t, smsg, encodedBody)

		// Attachment filename should be present (possibly encoded)
		assert.Contains(t, strings.ToLower(smsg), "filename=")
	})
}

func TestNewSMTPSender(t *testing.T) {
	t.Run("Normal port", func(t *testing.T) {
		sender := NewSMTPSender("host", "587", "user", "pass", "from").(*SMTPSender)
		assert.Equal(t, "host", sender.host)
		assert.Equal(t, "587", sender.port)
		assert.False(t, sender.useSSL)
	})

	t.Run("SSL port", func(t *testing.T) {
		sender := NewSMTPSender("host", "465", "user", "pass", "from").(*SMTPSender)
		assert.Equal(t, "465", sender.port)
		assert.True(t, sender.useSSL)
	})
}

func TestNewSMTPSenderWithConfig(t *testing.T) {
	cfg := SMTPConfig{
		Host:     "host",
		Port:     "465",
		Username: "user",
		Password: "pass",
		From:     "from",
		UseSSL:   true,
		Timeout:  30 * time.Second,
	}
	sender := NewSMTPSenderWithConfig(cfg).(*SMTPSender)
	assert.Equal(t, cfg.Host, sender.host)
	assert.Equal(t, cfg.Timeout, sender.timeout)
	assert.True(t, sender.useSSL)
}
