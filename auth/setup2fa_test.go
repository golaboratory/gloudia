package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup2FA_LongAccountNameReturnsErrorNotPanic(t *testing.T) {
	// QR コードに収まらない長い accountName は、QR エンコード失敗となる。
	// 旧実装はエラーを破棄して nil 画像を png.Encode に渡し panic していた。
	// 修正後はエラーを返し panic しないこと。
	longAccount := strings.Repeat("a", 4000)
	assert.NotPanics(t, func() {
		_, err := Setup2FA("Issuer", longAccount)
		assert.Error(t, err)
	})
}

func TestSetup2FA_Normal(t *testing.T) {
	resp, err := Setup2FA("Issuer", "user@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Secret)
	assert.Contains(t, resp.QRCodeB64, "data:image/png;base64,")
}
