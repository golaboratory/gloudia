package auth

import (
	"encoding/hex"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/newmo-oss/ergo"
)

var (
	// ErrInvalidKeySize は鍵のサイズが不正な場合のエラーです。
	ErrInvalidKeySize = ergo.NewSentinel("invalid key size: must be 32 bytes (64 hex characters)")
	// ErrInvalidHexKey はHexキーのデコードに失敗した場合のエラーです。
	ErrInvalidHexKey = ergo.NewSentinel("invalid hex key")
	// ErrSymmetricKeyCreationFailed は対称鍵の生成に失敗した場合のエラーです。
	ErrSymmetricKeyCreationFailed = ergo.NewSentinel("failed to create symmetric key")
	// ErrTokenVerificationFailed はトークンの検証に失敗した場合のエラーです。
	ErrTokenVerificationFailed = ergo.NewSentinel("failed to verify token")
	// ErrInvalidTokenPayloadUserID はトークンペイロードのuser_idが不正な場合のエラーです。
	ErrInvalidTokenPayloadUserID = ergo.NewSentinel("invalid token payload: user_id")
	// ErrInvalidTokenPayloadTenantID はトークンペイロードのtenant_idが不正な場合のエラーです。
	ErrInvalidTokenPayloadTenantID = ergo.NewSentinel("invalid token payload: tenant_id")
	// ErrInvalidTokenPayloadRoleID はトークンペイロードのrole_idが不正な場合のエラーです。
	ErrInvalidTokenPayloadRoleID = ergo.NewSentinel("invalid token payload: role_id")
)

// Claims はトークンに含まれるペイロード情報を定義します。
type Claims struct {
	UserID   int64  `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RoleID   int64  `json:"role_id"`
	// TokenType はトークンの発行元種別（例: "system_admin"）を表します。
	// 未設定（通常トークン・旧トークン）の場合は空文字となります。
	// 受信側ミドルウェアでクロス特権アクセス（別種トークンの使い回し）を拒否するために使用します。
	TokenType string `json:"token_type"`
}

// TokenMaker は PASETO トークンの生成と検証を行う構造体です。
type TokenMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

// NewTokenMaker は Hexエンコードされた32バイトの秘密鍵から TokenMaker を生成します。
// 鍵は必ず環境変数など安全な場所から供給してください。
func NewTokenMaker(hexKey string) (*TokenMaker, error) {
	if len(hexKey) != 64 { // 32 bytes * 2 (hex)
		return nil, ErrInvalidKeySize
	}

	bytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidHexKey, err)
	}

	key, err := paseto.V4SymmetricKeyFromBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSymmetricKeyCreationFailed, err)
	}

	return &TokenMaker{
		symmetricKey: key,
	}, nil
}

// CreateToken はユーザー情報を受け取り、署名・暗号化された PASETO トークン文字列を生成します。
// トークン種別（TokenType）は未設定（空文字）となります。発行元種別を付与する場合は
// CreateTokenWithType を使用してください。
func (maker *TokenMaker) CreateToken(userID int64, tenantID string, roleID int64, duration time.Duration) (string, error) {
	return maker.CreateTokenWithType(userID, tenantID, roleID, "", duration)
}

// CreateTokenWithType は発行元種別（tokenType）を埋め込んだ PASETO トークンを生成します。
// tokenType は受信側ミドルウェアでトークンの発行元（例: "system_admin"）を検証するために
// 使用します。tokenType が空文字の場合は種別クレームを付与しません（CreateToken と同等）。
func (maker *TokenMaker) CreateTokenWithType(userID int64, tenantID string, roleID int64, tokenType string, duration time.Duration) (string, error) {
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
	if tokenType != "" {
		token.Set("token_type", tokenType)
	}

	// v4.local (共有鍵) で暗号化
	encrypted := token.V4Encrypt(maker.symmetricKey, nil)
	return encrypted, nil
}

// VerifyToken はトークン文字列を復号・検証し、クレーム情報を返します。
func (maker *TokenMaker) VerifyToken(tokenString string) (*Claims, error) {
	claims, _, err := maker.VerifyTokenWithExpiry(tokenString)
	return claims, err
}

// VerifyTokenWithExpiry は VerifyToken と同様にトークンを検証し、クレームに加えて
// トークンの有効期限（exp クレーム）も返します。
// 有効期限を超えて維持される長寿命の接続（WebSocket 等）で、接続側に有効期限を
// 強制させるために使用します。exp が取得できない場合はゼロ値を返します。
func (maker *TokenMaker) VerifyTokenWithExpiry(tokenString string) (*Claims, time.Time, error) {
	parser := paseto.NewParser()

	// 有効期限などの標準ルールを検証に追加
	parser.AddRule(paseto.NotExpired())
	parser.AddRule(paseto.ValidAt(time.Now()))

	// 復号と解析
	token, err := parser.ParseV4Local(maker.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("%w: %v", ErrTokenVerificationFailed, err)
	}

	// クレームの抽出
	payload := &Claims{}

	// user_id (数値として取得)
	// 注意: go-paseto は JSON の数値を float64 として扱う場合があるため、適切にキャストします
	if err := token.Get("user_id", &payload.UserID); err != nil {
		// 文字列として入っている場合のフォールバックなどを検討しても良いですが、
		// ここでは厳密に型チェックします。
		return nil, time.Time{}, ErrInvalidTokenPayloadUserID
	}

	if err := token.Get("tenant_id", &payload.TenantID); err != nil {
		return nil, time.Time{}, ErrInvalidTokenPayloadTenantID
	}

	if err := token.Get("role_id", &payload.RoleID); err != nil {
		return nil, time.Time{}, ErrInvalidTokenPayloadRoleID
	}

	// token_type は任意クレーム（通常トークン・旧トークンには存在しない）。
	// 取得できない場合は空文字のままとし、後方互換性を維持する。
	_ = token.Get("token_type", &payload.TokenType)

	// 有効期限の取得（設定されていない場合はゼロ値）
	exp, expErr := token.GetExpiration()
	if expErr != nil {
		return payload, time.Time{}, nil
	}

	return payload, exp, nil
}

// Helper: 開発用などでランダムなHexキーを生成したい場合に使用
func GenerateRandomKey() string {
	key := paseto.NewV4SymmetricKey()
	return key.ExportHex()
}
