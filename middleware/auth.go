package middleware

import (
	"log/slog"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/golaboratory/gloudia/auth"
)

// NewAuthProvider は PASETO トークンを検証する Huma ミドルウェアを生成します。
// パターン: Huma (api.UseMiddleware で適用)
//
// Authorization ヘッダーから Bearer トークンを読み取り、検証を行います。
// 検証に成功した場合、トークンのペイロード（Claims）をコンテキストに保存します。
// ヘッダーが存在しない場合は 401 Unauthorized を返します。
//
// 引数:
//
//	maker: トークン検証に使用する TokenMaker インスタンス
func NewAuthProvider(maker *auth.TokenMaker) func(huma.Context, func(huma.Context)) {
	return newAuthProvider(maker, "")
}

// NewAuthProviderWithType は NewAuthProvider に加えて、トークンの発行元種別
// （Claims.TokenType）が expectedType と一致することを検証する Huma ミドルウェアを生成します。
// expectedType が空文字の場合は種別検証を行いません（NewAuthProvider と同等）。
//
// 用途: 複数の API が同一の PASETO 鍵を共有する構成で、ある API（例: system-admin-api）が
// 別種のトークン（テナント職員・ゲスト等）を受理してしまうクロス特権アクセスを防ぐために、
// expectedType="system_admin" のように発行元種別を必須化します。
func NewAuthProviderWithType(maker *auth.TokenMaker, expectedType string) func(huma.Context, func(huma.Context)) {
	return newAuthProvider(maker, expectedType)
}

// newAuthProvider は認証ミドルウェアの実体です。expectedType が非空のときのみ
// トークン種別検証を行います。
func newAuthProvider(maker *auth.TokenMaker, expectedType string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// 1. Authorization ヘッダーの取得
		authHeader := ctx.Header("Authorization")

		// ヘッダーがない場合は 401 Unauthorized を返す
		if authHeader == "" {
			ctx.SetStatus(401)
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

		// 3.5 トークン種別の検証（クロス特権アクセスの防止）
		// expectedType が指定されている API（例: system-admin-api）では、
		// 別種トークン（テナント職員・ゲスト等。TokenType が一致しない）を 403 で拒否する。
		if expectedType != "" && claims.TokenType != expectedType {
			slog.WarnContext(ctx.Context(), "token type mismatch",
				slog.String("expected_type", expectedType),
				slog.String("claims_token_type", claims.TokenType))
			ctx.SetStatus(403)
			return
		}

		// 4. テナント整合性チェック (テナント分離 / クロステナント防止)
		// テナント解決ミドルウェア (NewTenantResolution) がリクエストのホスト名から
		// 導出した tenant_id が Context に存在する場合、それが認証済みクレームの
		// TenantID と一致することを検証する。
		// 一致しない場合は「テナントAのトークンでテナントBのホストにアクセスする」
		// 越境アクセスの試行とみなし、403 Forbidden で拒否する。
		if hostTenant, ok := ctx.Context().Value(KeyTenantID).(string); ok && hostTenant != "" && hostTenant != claims.TenantID {
			slog.WarnContext(ctx.Context(), "tenant mismatch between host and authenticated claims",
				slog.String("host_tenant_id", hostTenant),
				slog.String("claims_tenant_id", claims.TenantID))
			ctx.SetStatus(403)
			return
		}

		// 5. 検証成功: Context に Claims (構造体) を保存
		// context_keys.go で定義した KeyClaims を使用
		ctx = huma.WithValue(ctx, KeyClaims, claims)

		// 6. 認証済みテナントを「信頼できる tenant_id」として Context に保存する。
		// これ以降のミドルウェア (RLS 等) やハンドラは、ホスト名由来の値ではなく
		// この検証済みの値を使用する。KeyTenantID は以後 authoritative な値となる。
		ctx = huma.WithValue(ctx, KeyTenantID, claims.TenantID)

		// 7. 次の処理へ
		next(ctx)
	}
}
