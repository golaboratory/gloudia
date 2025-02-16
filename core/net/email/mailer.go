package email

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// Server は、メールサーバーの設定を保持する構造体です。
type Server struct {
	Host         string // メールサーバーのホスト名
	Port         int    // メールサーバーのポート番号
	UseSsl       bool   // SSLを使用するかどうか
	NeedSmtpAuth bool   // SMTP認証が必要かどうか
}

// Credentials は、メールサーバーの認証情報を保持する構造体です。
type Credentials struct {
	Username string // ユーザー名
	Password string // パスワード
}

// Mailer は、メールの送信に必要な情報を保持する構造体です。
type Mailer struct {
	Server      Server      // メールサーバーの設定
	Credentials Credentials // 認証情報

	From        string   // 送信者のメールアドレス
	SenderName  string   // 送信者の名前
	To          []string // 宛先のメールアドレス
	Cc          []string // CCのメールアドレス
	Bcc         []string // BCCのメールアドレス
	Subject     string   // メールの件名
	Body        string   // メールの本文
	AttachFiles []string // 添付ファイルのパス

	UseHTMLBody bool // HTML形式の本文を使用するかどうか
}

// Send は、メールを送信するメソッドです。
// メールの送信に成功した場合は nil を返し、エラーが発生した場合はエラーを返します。
func (m *Mailer) Send() error {
	e := email.NewEmail()

	from := fmt.Sprintf("%s <%s>", m.SenderName, m.From)
	if m.SenderName == "" {
		from = m.From
	}
	e.From = from
	e.To = m.To
	e.Cc = m.Cc
	e.Bcc = m.Bcc
	e.Subject = m.Subject

	if m.UseHTMLBody {
		e.HTML = []byte(m.Body)
	} else {
		e.Text = []byte(m.Body)
	}

	for _, file := range m.AttachFiles {
		_, err := e.AttachFile(file)
		if err != nil {
			return err
		}
	}

	var auth smtp.Auth = nil
	if m.Server.NeedSmtpAuth {
		auth = smtp.PlainAuth("", m.Credentials.Username, m.Credentials.Password, m.Server.Host)
	}
	err := e.Send(fmt.Sprintf("%s:%d", m.Server.Host, m.Server.Port), auth)

	if err != nil {
		return err
	}

	return nil
}
