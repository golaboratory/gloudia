package config

// Configure は設定情報をロードするためのインターフェースです。
type Configure interface {
	// Load は設定情報を読み込みます。
	Load() error
}
