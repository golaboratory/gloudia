package mail

// Sender はメール送信のインターフェースです
type Sender interface {
	// SendEmail はメールを送信します。
	// subject は件名、content は本文、to / cc / bcc は宛先メールアドレス、
	// attachFiles は添付するファイルのパスです。
	// 空文字・空白のみのアドレスは除外されます。to+cc+bcc を合わせて
	// 少なくとも 1 件の宛先が必要です。
	SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error
}
