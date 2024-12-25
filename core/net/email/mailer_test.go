package email

import (
	"github.com/stretchr/testify/assert"
	"net/smtp"
	"testing"
)

// Mock for smtp.SendMail
var sendMail = smtp.SendMail

func TestMailer_Send(t *testing.T) {
	// Mock smtp.SendMail
	smtp.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}
	defer func() { smtp.SendMail = sendMail }()

	tests := []struct {
		name    string
		mailer  Mailer
		wantErr bool
	}{
		{
			name: "successful send",
			mailer: Mailer{
				Server: Server{
					Host: "smtp.example.com",
					Port: 587,
				},
				Credentials: Credentials{
					Username: "user",
					Password: "pass",
				},
				From:        "from@example.com",
				SenderName:  "Sender",
				To:          []string{"to@example.com"},
				Subject:     "Test Subject",
				Body:        "Test Body",
				UseHTMLBody: false,
			},
			wantErr: false,
		},
		{
			name: "send with attachment",
			mailer: Mailer{
				Server: Server{
					Host: "smtp.example.com",
					Port: 587,
				},
				Credentials: Credentials{
					Username: "user",
					Password: "pass",
				},
				From:        "from@example.com",
				SenderName:  "Sender",
				To:          []string{"to@example.com"},
				Subject:     "Test Subject",
				Body:        "Test Body",
				AttachFiles: []string{"testfile.txt"},
				UseHTMLBody: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mailer.Send()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
