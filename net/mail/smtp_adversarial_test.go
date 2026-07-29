package mail

import (
	"bufio"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSMTPServer accepts a single SMTP session over a non-TLS connection and
// drives the minimal happy-path protocol (no AUTH, no STARTTLS). It records the
// MAIL FROM / RCPT TO commands and the DATA body so the test can assert them.
func fakeSMTPServer(t *testing.T) (host string, port string, got *struct {
	mailFrom string
	rcpts    []string
	data     string
}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	rec := &struct {
		mailFrom string
		rcpts    []string
		data     string
	}{}

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		defer ln.Close()

		r := bufio.NewReader(conn)
		w := func(s string) { _, _ = conn.Write([]byte(s)) }

		w("220 fake ESMTP\r\n")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.TrimSpace(line)
			upper := strings.ToUpper(cmd)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				w("250 ok\r\n")
			case strings.HasPrefix(upper, "MAIL FROM"):
				rec.mailFrom = cmd
				w("250 ok\r\n")
			case strings.HasPrefix(upper, "RCPT TO"):
				rec.rcpts = append(rec.rcpts, cmd)
				w("250 ok\r\n")
			case strings.HasPrefix(upper, "DATA"):
				w("354 end data with <CR><LF>.<CR><LF>\r\n")
				var b strings.Builder
				for {
					dl, err := r.ReadString('\n')
					if err != nil {
						break
					}
					if strings.TrimRight(dl, "\r\n") == "." {
						break
					}
					b.WriteString(dl)
				}
				rec.data = b.String()
				w("250 ok\r\n")
			case strings.HasPrefix(upper, "QUIT"):
				w("221 bye\r\n")
				return
			default:
				w("250 ok\r\n")
			}
		}
	}()

	h, p, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	return h, p, rec
}

func TestSend_HappyPathDeliversToFakeServer(t *testing.T) {
	host, port, rec := fakeSMTPServer(t)

	s := &SMTPSender{
		host:    host,
		port:    port,
		from:    "from@example.com",
		timeout: 5 * time.Second,
		// username == "" -> AUTH skipped; useSSL == false -> plain TCP.
	}

	err := s.send([]string{"a@example.com", "b@example.com"}, []byte("Subject: hi\r\n\r\nbody"))
	require.NoError(t, err)

	// Allow the goroutine to finish recording the data before asserting.
	deadline := time.Now().Add(2 * time.Second)
	for rec.data == "" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	assert.Contains(t, rec.mailFrom, "<from@example.com>")
	assert.Len(t, rec.rcpts, 2)
	assert.Contains(t, fmt.Sprint(rec.rcpts), "<a@example.com>")
	assert.Contains(t, fmt.Sprint(rec.rcpts), "<b@example.com>")
	assert.Contains(t, rec.data, "body")
}

func TestSendEmail_NoRecipients(t *testing.T) {
	s := &SMTPSender{from: "from@example.com", host: "127.0.0.1", port: "2525", timeout: time.Second}

	err := s.SendEmail("subj", "body", []string{"   "}, []string{""}, nil, nil)
	assert.ErrorIs(t, err, ErrNoRecipientsSpecified)
}

func TestBuildMessage_MissingAttachmentReturnsError(t *testing.T) {
	s := &SMTPSender{from: "from@example.com"}
	missing := filepath.Join(t.TempDir(), "does-not-exist.txt")

	msg, err := s.buildMessage("subj", "body", []string{"to@example.com"}, nil, []string{missing})
	require.Error(t, err)
	assert.Nil(t, msg)
}

func TestEncodeAddress_PreservesEmailWhenNameIsJapanese(t *testing.T) {
	s := &SMTPSender{}

	out := s.encodeAddress("管理者 <admin@example.com>")
	assert.Contains(t, out, "=?utf-8?", "display name should be RFC2047 encoded")
	assert.Contains(t, out, "<admin@example.com>", "email address part must be left intact and unencoded")
	assert.NotContains(t, out, "管理者", "raw Japanese display name must not leak unencoded")
}
