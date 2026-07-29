package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptor(t *testing.T) {
	// 32バイトのランダムキー生成
	key := make([]byte, 32)
	rand.Read(key)

	cryptor, err := NewCryptor(key)
	assert.NoError(t, err)

	t.Run("Encrypt and Decrypt", func(t *testing.T) {
		plain := "This is a secret message. 秘密情報です。"

		// 暗号化
		encrypted, err := cryptor.Encrypt(plain)
		assert.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, plain, encrypted)

		// 復号
		decrypted, err := cryptor.Decrypt(encrypted)
		assert.NoError(t, err)
		assert.Equal(t, plain, decrypted)
	})

	t.Run("Decrypt invalid input", func(t *testing.T) {
		// Base64として不正
		_, err := cryptor.Decrypt("invalid-base64")
		assert.Error(t, err)

		// 短すぎる (Nonceサイズ未満)
		short := hex.EncodeToString([]byte("short"))
		_, err = cryptor.Decrypt(short)
		assert.Error(t, err)

		// 改ざんされたデータ (鍵が合わない/タグ不整合)
		// 正しい暗号文の一部を変更
		valid, _ := cryptor.Encrypt("test")
		// Base64デコード -> 1バイト変更 -> エンコード は手間なので、別の鍵で復号を試みる

		key2 := make([]byte, 32)
		rand.Read(key2) // 別の鍵
		cryptor2, _ := NewCryptor(key2)

		_, err = cryptor2.Decrypt(valid)
		assert.Error(t, err) // Authentication failed
	})

	t.Run("Invalid Key Size", func(t *testing.T) {
		shortKey := []byte("too short")
		_, err := NewCryptor(shortKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be 32 bytes")
	})
}
