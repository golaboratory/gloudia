package email

import (
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/assert"
)

func TestMailer_Send(t *testing.T) {
	mockServer := smtpmock.New(smtpmock.ConfigurationAttr{})
	defer mockServer.Stop()

	// テスト終了後にSMTPサーバ停止
	t.Cleanup(func() {
		if err := mockServer.Stop(); err != nil {
			t.Log(err)
		}
	})

	// SMTPサーバ起動
	if err := mockServer.Start(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		mailer  Mailer
		wantErr bool
	}{
		{
			name: "successful send",
			mailer: Mailer{
				Server: Server{
					Host:         "localhost",
					Port:         mockServer.PortNumber(),
					NeedSmtpAuth: false,
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
					Host:         "localhost",
					Port:         mockServer.PortNumber(),
					NeedSmtpAuth: false,
				},
				Credentials: Credentials{
					Username: "user",
					Password: "pass",
				},
				From:       "from@example.com",
				SenderName: "Sender",
				To:         []string{"to@example.com"},
				Subject:    "Test Subject",
				Body:       "Test Body",
				AttachFiles: []string{
					"../../../testdata/core/email/test.jpg",
					"../../../testdata/core/email/testtext.txt"},
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
