package mail

// Sender はメール送信のインターフェースです
type Sender interface {
	SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error
}
