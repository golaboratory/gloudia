package mail

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/newmo-oss/ergo"
)

// SMTPSender は SMTP を使用してメールを送信する Sender インターフェースの実装です。
// 日本語エンコーディング、添付ファイル、SSL/TLS、STARTTLS に対応しています。
type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
	useSSL   bool
	insecure bool
	timeout  time.Duration
}

// SMTPConfig は SMTP 送信者の詳細な設定を保持する構造体です。
type SMTPConfig struct {
	// Host は SMTP サーバーのホスト名です。
	Host string
	// Port は SMTP サーバーのポート番号です（例: "465", "587"）。
	Port string
	// Username は認証に使用するユーザー名です。
	Username string
	// Password は認証に使用するパスワードです。
	Password string
	// From は送信元のアドレスです（例: "名前 <admin@example.com>"）。
	From string
	// UseSSL が true の場合、接続の最初から SSL/TLS を使用します（主にポート 465）。
	UseSSL bool
	// Insecure が true の場合、サーバーの TLS 証明書の検証をスキップします。
	// 本番環境では false にすることを強く推奨します。
	Insecure bool
	// Timeout は接続および各コマンドのタイムアウト時間です。指定しない場合は 10秒となります。
	Timeout time.Duration
}

// NewSMTPSender は従来のパラメータ形式で SMTPSender を作成します。
// ポート番号が "465" の場合は、自動的に SSL を有効にします。
func NewSMTPSender(host, port, username, password, from string) Sender {
	s := &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		timeout:  10 * time.Second,
	}
	if port == "465" {
		s.useSSL = true
	}
	return s
}

// NewSMTPSenderWithConfig は SMTPConfig 構造体を使用して SMTPSender を作成します。
func NewSMTPSenderWithConfig(cfg SMTPConfig) Sender {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &SMTPSender{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		useSSL:   cfg.UseSSL,
		insecure: cfg.Insecure,
		timeout:  cfg.Timeout,
	}
}

// SendEmail は SMTP を使用してメールを送信します。
// 宛先 (to, cc, bcc) や添付ファイル (attachFiles) をサポートし、
// 日本語の件名や名前は RFC 2047 形式で自動的にエンコードされます。
func (s *SMTPSender) SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error {
	to = s.filterEmpty(to)
	cc = s.filterEmpty(cc)
	bcc = s.filterEmpty(bcc)

	if len(to) == 0 && len(cc) == 0 && len(bcc) == 0 {
		return ergo.New("no recipients specified")
	}

	// メッセージの構築
	message, err := s.buildMessage(subject, content, to, cc, attachFiles)
	if err != nil {
		return ergo.New("failed to build email", slog.String("error", err.Error()))
	}

	// RCPT TO コマンドで使用する全受信者のリストを作成
	allRecipients := s.uniqueRecipients(to, cc, bcc)

	return s.send(allRecipients, message)
}

// buildMessage は MIME 形式のメッセージデータを構築します。
func (s *SMTPSender) buildMessage(subject, content string, to, cc, attachFiles []string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	// RFC 5322 ヘッダー
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.encodeAddress(s.from)))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", s.encodeAddressList(to)))
	if len(cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", s.encodeAddressList(cc)))
	}
	// 件名の日本語エンコード
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.BEncoding.Encode("utf-8", subject)))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("MIME-Version: 1.0\r\n")

	contentType := "text/plain"
	if s.isHTML(content) {
		contentType = "text/html"
	}

	if len(attachFiles) == 0 {
		// 添付ファイルがない場合のシングルパート
		buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\r\n", contentType))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		buf.WriteString(s.base64Encode(content))
		buf.WriteString("\r\n")
		return buf.Bytes(), nil
	}

	// 添付ファイルがある場合のマルチパート
	writer := multipart.NewWriter(buf)
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", writer.Boundary()))

	// 本文パート
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", contentType))
	h.Set("Content-Transfer-Encoding", "base64")
	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}
	_, _ = part.Write([]byte(s.base64Encode(content)))

	// 添付ファイルパート
	for _, file := range attachFiles {
		if err := s.attachFile(writer, file); err != nil {
			return nil, err
		}
	}

	_ = writer.Close()
	return buf.Bytes(), nil
}

// attachFile は指定されたファイルをマルチパートライターに追加します。
func (s *SMTPSender) attachFile(w *multipart.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	name := filepath.Base(filename)
	h := make(textproto.MIMEHeader)

	ext := filepath.Ext(name)
	mtype := mime.TypeByExtension(ext)
	if mtype == "" {
		mtype = "application/octet-stream"
	}

	// ファイル名の日本語エンコード
	encodedName := mime.BEncoding.Encode("utf-8", name)
	h.Set("Content-Type", fmt.Sprintf("%s; name=\"%s\"", mtype, encodedName))
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", encodedName))
	h.Set("Content-Transfer-Encoding", "base64")

	part, err := w.CreatePart(h)
	if err != nil {
		return err
	}

	encoder := base64.NewEncoder(base64.StdEncoding, part)
	if _, err := io.Copy(encoder, file); err != nil {
		return err
	}
	return encoder.Close()
}

// send は SMTP サーバーへの接続、認証、およびデータの送信を行います。
func (s *SMTPSender) send(recipients []string, message []byte) error {
	addr := net.JoinHostPort(s.host, s.port)

	var client *smtp.Client
	var err error

	if s.useSSL {
		// 暗黙的な SSL/TLS 接続 (例: ポート 465)
		tlsConfig := &tls.Config{
			ServerName:         s.host,
			InsecureSkipVerify: s.insecure,
		}
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: s.timeout}, "tcp", addr, tlsConfig)
		if err != nil {
			return ergo.New("SSL connection failed", slog.String("error", err.Error()))
		}
		client, err = smtp.NewClient(conn, s.host)
		if err != nil {
			return ergo.New("SMTP client creation failed", slog.String("error", err.Error()))
		}
	} else {
		// 標準的な TCP 接続と STARTTLS によるアップグレード
		conn, err := net.DialTimeout("tcp", addr, s.timeout)
		if err != nil {
			return ergo.New("TCP connection failed", slog.String("error", err.Error()))
		}
		client, err = smtp.NewClient(conn, s.host)
		if err != nil {
			return ergo.New("SMTP client creation failed", slog.String("error", err.Error()))
		}

		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName:         s.host,
				InsecureSkipVerify: s.insecure,
			}
			if err = client.StartTLS(tlsConfig); err != nil {
				return ergo.New("STARTTLS failed", slog.String("error", err.Error()))
			}
		}
	}
	defer client.Quit()

	// 認証
	if s.username != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err = client.Auth(auth); err != nil {
			return ergo.New("SMTP authentication failed", slog.String("error", err.Error()))
		}
	}

	// トランザクション開始
	if err = client.Mail(s.from); err != nil {
		return ergo.New("MAIL FROM failed", slog.String("error", err.Error()))
	}
	for _, rec := range recipients {
		if err = client.Rcpt(rec); err != nil {
			return ergo.New("RCPT TO failed for %s", slog.String("error", err.Error()))
		}
	}

	// 送信データ書き込み
	w, err := client.Data()
	if err != nil {
		return ergo.New("DATA command failed", slog.String("error", err.Error()))
	}
	if _, err = w.Write(message); err != nil {
		return ergo.New("failed to write message", slog.String("error", err.Error()))
	}
	if err = w.Close(); err != nil {
		return ergo.New("failed to close DATA writer", slog.String("error", err.Error()))
	}

	return nil
}

// isHTML はコンテンツが HTML かどうかを判定します。
func (s *SMTPSender) isHTML(content string) bool {
	c := strings.ToLower(strings.TrimSpace(content))
	return strings.HasPrefix(c, "<html") || strings.HasPrefix(c, "<!doctype html")
}

// base64Encode は文字列を Base64 エンコードします。
func (s *SMTPSender) base64Encode(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

// encodeAddress は "表示名 <email@example.com>" 形式のアドレスの表示名部分に
// 日本語が含まれる場合、適切にエンコードします。
func (s *SMTPSender) encodeAddress(addr string) string {
	if !s.hasNonASCII(addr) {
		return addr
	}
	if !strings.Contains(addr, "<") {
		return mime.BEncoding.Encode("utf-8", addr)
	}
	parts := strings.SplitN(addr, "<", 2)
	name := strings.TrimSpace(parts[0])
	rest := parts[1]
	return fmt.Sprintf("%s <%s", mime.BEncoding.Encode("utf-8", name), rest)
}

// encodeAddressList は複数のアドレスをエンコードしてカンマで連結します。
func (s *SMTPSender) encodeAddressList(list []string) string {
	encoded := make([]string, len(list))
	for i, addr := range list {
		encoded[i] = s.encodeAddress(addr)
	}
	return strings.Join(encoded, ", ")
}

// hasNonASCII は文字列に非 ASCII 文字が含まれているか判定します。
func (s *SMTPSender) hasNonASCII(val string) bool {
	for i := 0; i < len(val); i++ {
		if val[i] > 127 {
			return true
		}
	}
	return false
}

// filterEmpty はスライスから空文字列を除去します。
func (s *SMTPSender) filterEmpty(slice []string) []string {
	var res []string
	for _, val := range slice {
		if t := strings.TrimSpace(val); t != "" {
			res = append(res, t)
		}
	}
	return res
}

// uniqueRecipients は重複を除去した全受信者のリストを作成します。
func (s *SMTPSender) uniqueRecipients(to, cc, bcc []string) []string {
	set := make(map[string]struct{})
	var res []string
	for _, slice := range [][]string{to, cc, bcc} {
		for _, r := range slice {
			if _, ok := set[r]; !ok {
				set[r] = struct{}{}
				res = append(res, r)
			}
		}
	}
	return res
}
