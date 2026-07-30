// Package mail は SMTP を使用したメール送信機能を提供します。
// SSL/TLS および STARTTLS 接続、日本語の件名・ファイル名エンコーディング（RFC 2047）、
// 添付ファイル付きマルチパートメールをサポートします。
//
// 注意: STARTTLS は日和見的（opportunistic）です。サーバーが STARTTLS を
// 広告しない場合、メールはエラーにならず平文のまま送信されます。
// 現状 TLS を必須化する手段はありません。
package mail
