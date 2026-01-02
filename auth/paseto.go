package auth

import (
	"encoding/hex"
	"log/slog"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/newmo-oss/ergo"
)

// Claims はトークンに含まれるペイロード情報を定義します。
type Claims struct {
	UserID   int64  `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RoleID   int64  `json:"role_id"`
}

// TokenMaker は PASETO トークンの生成と検証を行う構造体です。
type TokenMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

// NewTokenMaker は Hexエンコードされた32バイトの秘密鍵から TokenMaker を生成します。
// 鍵は必ず環境変数など安全な場所から供給してください。
func NewTokenMaker(hexKey string) (*TokenMaker, error) {
	if len(hexKey) != 64 { // 32 bytes * 2 (hex)
		return nil, ergo.New("invalid key size: must be 32 bytes (64 hex characters)")
	}

	bytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, ergo.New("invalid hex key", slog.String("error", err.Error()))
	}

	key, err := paseto.V4SymmetricKeyFromBytes(bytes)
	if err != nil {
		return nil, ergo.New("failed to create symmetric key", slog.String("error", err.Error()))
	}

	return &TokenMaker{
		symmetricKey: key,
	}, nil
}

// CreateToken はユーザー情報を受け取り、署名・暗号化された PASETO トークン文字列を生成します。
func (maker *TokenMaker) CreateToken(userID int64, tenantID string, roleID int64, duration time.Duration) (string, error) {
	token := paseto.NewToken()

	// 標準クレームの設定
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(duration))

	// カスタムクレームの設定 (JSONとしてシリアライズ可能な型を渡す)
	// ※ int64はJSONでは数値ですが、Pasetoライブラリの仕様に合わせて文字列化するか、SetString等を使うか選択します。
	// ここでは汎用的な Set メソッドを使用します。
	token.Set("user_id", userID)
	token.Set("tenant_id", tenantID)
	token.Set("role_id", roleID)

	// v4.local (共有鍵) で暗号化
	encrypted := token.V4Encrypt(maker.symmetricKey, nil)
	return encrypted, nil
}

// VerifyToken はトークン文字列を復号・検証し、クレーム情報を返します。
func (maker *TokenMaker) VerifyToken(tokenString string) (*Claims, error) {
	parser := paseto.NewParser()

	// 有効期限などの標準ルールを検証に追加
	parser.AddRule(paseto.NotExpired())
	parser.AddRule(paseto.ValidAt(time.Now()))

	// 復号と解析
	token, err := parser.ParseV4Local(maker.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, ergo.New("failed to verify token", slog.String("error", err.Error()))
	}

	// クレームの抽出
	payload := &Claims{}

	// user_id (数値として取得)
	// 注意: go-paseto は JSON の数値を float64 として扱う場合があるため、適切にキャストします
	if err := token.Get("user_id", &payload.UserID); err != nil {
		// 文字列として入っている場合のフォールバックなどを検討しても良いですが、
		// ここでは厳密に型チェックします。
		return nil, ergo.New("invalid token payload: user_id")
	}

	if err := token.Get("tenant_id", &payload.TenantID); err != nil {
		return nil, ergo.New("invalid token payload: tenant_id")
	}

	if err := token.Get("role_id", &payload.RoleID); err != nil {
		return nil, ergo.New("invalid token payload: role_id")
	}

	return payload, nil
}

// Helper: 開発用などでランダムなHexキーを生成したい場合に使用
func GenerateRandomKey() string {
	key := paseto.NewV4SymmetricKey()
	return key.ExportHex()
}
