package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex" // 追加
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Cryptography はAES-GCMに基づく暗号化および復号化機能を提供する型です。
type Cryptography struct {
	// 必要に応じてフィールドを追加
}

// EncryptString は与えられた文字列dataStringをAES-GCMを用いて暗号化します。
// 内部的にはNewSecurityConfigにより取得したパスワードとソルトから鍵を生成し使用します。
// 引数:
//   - dataString: 暗号化対象の文字列
//
// 戻り値:
//   - string: Base64エンコードされた暗号化済み文字列
//   - error: エラー情報
func (c *Cryptography) EncryptString(dataString string) (string, error) {

	conf := NewSecurityConfig()

	key := pbkdf2.Key([]byte(conf.Password), []byte(conf.Salt), 1000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext := []byte(dataString)

	iv := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// 暗号化処理
	encryptedData := aead.Seal(iv, iv, plaintext, nil)
	encoded := base64.StdEncoding.EncodeToString(encryptedData)
	return encoded, nil
}

// DecryptString は与えられたBase64エンコード済みの暗号化文字列dataStringをAES-GCMを用いて復号化します。
// 内部的にはNewSecurityConfigにより取得したパスワードとソルトから鍵を生成し使用します。
// 引数:
//   - dataString: EncryptStringの出力であるBase64エンコードされた暗号化済み文字列
//
// 戻り値:
//   - string: 復号化された文字列
//   - error: エラー情報
func (c *Cryptography) DecryptString(dataString string) (string, error) {

	conf := NewSecurityConfig()

	key := pbkdf2.Key([]byte(conf.Password), []byte(conf.Salt), 1000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	iv := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	encryptedData, err := base64.StdEncoding.DecodeString(dataString)
	if err != nil {
		return "", err
	}

	if len(encryptedData) < aead.NonceSize() {
		return "", err
	}

	iv2, ciphertext := encryptedData[:aead.NonceSize()], encryptedData[aead.NonceSize():]
	plaintext2, err := aead.Open(nil, iv2, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext2), nil
}

// HashString は与えられた文字列dataをSecurityConfigに設定されたソルトを利用しpbkdf2で処理することで
// 復号不可能なハッシュ値（16進数文字列）に変換します。
// 同じ文字列を引数として渡した場合は常に同じ結果を返します。
// 引数:
//   - data: ハッシュ化対象の文字列
//
// 戻り値:
//   - string: 16進数文字列のハッシュ値
func (c *Cryptography) HashString(data string) string {
	conf := NewSecurityConfig()
	// pbkdf2 を利用し、固定のソルトと繰り返し回数でハッシュ値を導出する（復号不可能）
	hash := pbkdf2.Key([]byte(data), []byte(conf.Salt), 1000, 32, sha256.New)
	return hex.EncodeToString(hash)
}
