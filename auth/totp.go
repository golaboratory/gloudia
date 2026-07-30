package auth

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"image/png"
	"time"

	"github.com/newmo-oss/ergo"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	// ErrTOTPKeyGenerationFailed はTOTP関連の生成処理（キー生成、QRコード画像の生成、
	// PNGエンコード、検証時のTOTPコード生成）に失敗した場合のエラーです。
	ErrTOTPKeyGenerationFailed = ergo.NewSentinel("failed to generate TOTP key")
)

// Setup2FAResponse は2要素認証（2FA）セットアップ時に返されるレスポンス構造体です。
// シークレットキーやQRコードの情報を含みます。
type Setup2FAResponse struct {
	// Secret はユーザーのDBに保存すべきシークレットキーです。
	// 本番環境ではクライアントに返さずサーバー側で保存する運用が推奨されます。
	Secret string `json:"secret" doc:"ユーザーのDBに保存すべきシークレットキー（本番ではクライアントに返さずサーバー側で保存）"`
	// QRCodeURI は "otpauth://" から始まるURI文字列です。
	QRCodeURI string `json:"qr_code_uri" doc:"otpauth://から始まるURI"`
	// QRCodeB64 は "data:image/png;base64," プレフィックス付きのデータURIです。
	// そのままHTMLのimgタグのsrc属性に指定して表示できます。
	QRCodeB64 string `json:"qr_code_base64" doc:"imgタグのsrcにそのまま指定できるPNG画像のデータURI（data:image/png;base64,...）"`
}

// Setup2FA は指定された発行者名とアカウント名を使用して新しいTOTPキーを生成し、
// QRコードを含むセットアップ情報を返します。
func Setup2FA(issuer string, accountName string) (*Setup2FAResponse, error) {

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, ergo.Wrap(ErrTOTPKeyGenerationFailed, err.Error())
	}

	// 2. 画像の生成とBase64化
	// フロントエンドで <img src="..."> と書けるようにバッファへ書き出す。
	// QR エンコードは accountName/issuer が長すぎる場合などに失敗する。エラーを
	// 破棄して nil 画像を png.Encode に渡すと nil ポインタ参照で panic するため、
	// 双方のエラーを確認する。
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, ergo.Wrap(ErrTOTPKeyGenerationFailed, err.Error())
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, ergo.Wrap(ErrTOTPKeyGenerationFailed, err.Error())
	}
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
//
// 注意（リプレイ保護なし）: totp.Validate はステートレスであり、現在/前後の
// タイムステップ（既定 skew=1, period=30 で約90秒間）に一致するコードを
// 何度でも受理します。同一コードの再利用（リプレイ）を防ぐには、呼び出し側で
// 「最後に受理したタイムステップ」をユーザーごとに保存し、それ以下のステップの
// コードを拒否する必要があります。そのためのステップ値は Verify2FAWithTimeStep を
// 使用して取得できます。
func Verify2FA(secret string, code string) bool {
	valid := totp.Validate(code, secret)
	return valid
}

// Verify2FAWithTimeStep は Verify2FA と同様にコードを検証しつつ、一致した
// タイムステップ（Unix時刻 / period の整数値）を返します。
// 呼び出し側はこの値を「最後に受理したステップ」として永続化し、
// step <= 最後に受理したステップ のコードを拒否することでリプレイを防止できます。
//
// 戻り値: (検証成功か, 一致したタイムステップ, エラー)。
// 検証失敗時は (false, 0, nil) を返します。
func Verify2FAWithTimeStep(secret string, code string) (bool, int64, error) {
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}
	now := time.Now()
	// totp.Validate と同等に現在ステップの前後 skew 個を検査する。
	for i := -int(opts.Skew); i <= int(opts.Skew); i++ {
		t := now.Add(time.Duration(i) * time.Duration(opts.Period) * time.Second)
		expected, err := totp.GenerateCodeCustom(secret, t, opts)
		if err != nil {
			return false, 0, ergo.Wrap(ErrTOTPKeyGenerationFailed, err.Error())
		}
		// タイミング攻撃を避けるため定数時間比較を使用する。
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true, t.Unix() / int64(opts.Period), nil
		}
	}
	return false, 0, nil
}
