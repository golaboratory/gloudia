package auth

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"log/slog"

	"github.com/newmo-oss/ergo"
	"github.com/pquerna/otp/totp"
)

// Setup2FAResponse は2要素認証（2FA）セットアップ時に返されるレスポンス構造体です。
// シークレットキーやQRコードの情報を含みます。
type Setup2FAResponse struct {
	// Secret はユーザーのDBに保存すべきシークレットキーです。
	// 本番環境ではクライアントに返さずサーバー側で保存する運用が推奨されます。
	Secret string `json:"secret" doc:"ユーザーのDBに保存すべきシークレットキー（本番ではクライアントに返さずサーバー側で保存）"`
	// QRCodeURI は "otpauth://" から始まるURI文字列です。
	QRCodeURI string `json:"qr_code_uri" doc:"otpauth://から始まるURI"`
	// QRCodeB64 はHTMLのimgタグで表示可能なBase64エンコードされたPNG画像データです。
	QRCodeB64 string `json:"qr_code_base64" doc:"HTMLのimgタグで表示可能なBase64画像データ"`
}

// Setup2FA は指定された発行者名とアカウント名を使用して新しいTOTPキーを生成し、
// QRコードを含むセットアップ情報を返します。
func Setup2FA(issuer string, accountName string) (*Setup2FAResponse, error) {

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, ergo.New("failed to generate TOTP key", slog.String("error", err.Error()))
	}

	// 2. 画像の生成とBase64化
	// フロントエンドで <img src="..."> と書けるようにバッファへ書き出す
	var buf bytes.Buffer
	img, _ := key.Image(200, 200)
	png.Encode(&buf, img)
	imgBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	// 3. レスポンス生成
	resp := &Setup2FAResponse{}
	resp.Secret = key.Secret() // 本来はここでDB保存を行う
	resp.QRCodeURI = key.String()
	resp.QRCodeB64 = "data:image/png;base64," + imgBase64

	return resp, nil

}

// Verify2FA は指定されたシークレットキーとワンタイムパスコード（code）を使用して
// 2要素認証の検証を行います。検証に成功した場合は true を返します。
func Verify2FA(secret string, code string) bool {
	valid := totp.Validate(code, secret)
	return valid
}
