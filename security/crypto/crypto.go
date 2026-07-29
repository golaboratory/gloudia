package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrInvalidKeyLength は鍵長が32バイトでない場合に返されるエラーです。
	ErrInvalidKeyLength = ergo.NewSentinel("key length must be 32 bytes for AES-256")
	// ErrCiphertextTooShort は暗号文がNonceサイズ未満の場合に返されるエラーです。
	ErrCiphertextTooShort = ergo.NewSentinel("ciphertext too short")
)

// Cryptor はデータの暗号化・復号を行う構造体です。
type Cryptor struct {
	gcm cipher.AEAD
}

// NewCryptor は指定された32バイト(AES-256用)の鍵でCryptorを初期化します。
// key は必ず安全な乱数、または適切に管理されたシークレットである必要があります。
func NewCryptor(key []byte) (*Cryptor, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to create cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ergo.Wrap(err, "failed to create gcm")
	}

	return &Cryptor{gcm: gcm}, nil
}

// Encrypt は平文文字列を暗号化し、Base64URLエンコードされた文字列を返します。
// 出力にはNonceが含まれます。
func (c *Cryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", ergo.Wrap(err, "failed to generate nonce")
	}

	// Seal: nonce + ciphertext + tag (GCM handles tag)
	// ここでは Nonce を暗号文の先頭に結合して保存します (Decrypt時に取り出す)
	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt は Encrypt で生成されたBase64URL文字列を復号し、平文を返します。
func (c *Cryptor) Decrypt(encryptedStr string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", ergo.Wrap(err, "failed to decode base64")
	}

	nonceSize := c.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ergo.Wrap(err, "failed to decrypt")
	}

	return string(plaintext), nil
}
