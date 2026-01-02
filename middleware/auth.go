package middleware

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/golaboratory/gloudia/auth"
)

// NewAuthProvider は PASETO トークンを検証する Huma ミドルウェアを生成します。
//
// Authorization ヘッダーから Bearer トークンを読み取り、検証を行います。
// 検証に成功した場合、トークンのペイロード（Claims）をコンテキストに保存します。
// ヘッダーが存在しない場合は検証をスキップし、後続の処理に委譲します（ゲストアクセスの考慮）。
//
// 引数:
//
//	maker: トークン検証に使用する TokenMaker インスタンス
func NewAuthProvider(maker *auth.TokenMaker) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// 1. Authorization ヘッダーの取得
		authHeader := ctx.Header("Authorization")

		// ヘッダーがない場合は検証をスキップして次の処理へ (ゲストアクセスの可能性)
		// ※ 認証必須のエンドポイントかどうかは、各Handler側またはHumaのSecurity Schemeで制御します
		if authHeader == "" {
			ctx.SetStatus(http.StatusUnauthorized)
			return
		}

		// 2. Bearer スキーマの検証
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
			// ヘッダーがあるのに形式が不正な場合は 401 を返す
			ctx.SetStatus(401)
			return
		}

		tokenString := fields[1]

		// 3. トークンの検証 (internal/auth パッケージ利用)
		claims, err := maker.VerifyToken(tokenString)
		if err != nil {
			// 期限切れや改ざん検知時
			ctx.SetStatus(401)
			return
		}

		// 4. 検証成功: Context に Claims (構造体) を保存
		// context_keys.go で定義した KeyClaims を使用
		ctx = huma.WithValue(ctx, KeyClaims, claims)

		// 5. 次の処理へ
		next(ctx)
	}
}
