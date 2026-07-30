// Package middleware はミドルウェアの2つのパターンを提供します。
//
// # Chi パターン (標準 net/http ミドルウェア)
//
//	func(http.Handler) http.Handler
//
// 対象: NewTenantResolution, NewLogger, NewRobotTag,
// NewCORS (*cors.Cors を返すため、その Handler メソッドを r.Use() に渡す)
// 使用場所: Chi ルーターの r.Use() で適用
//
// # Huma パターン (Huma API ミドルウェア)
//
//	func(huma.Context, func(huma.Context))
//
// 対象: NewAuthProvider, NewAuthProviderWithType, NewRLSProvider, NewRedisRateLimiter
// 使用場所: Huma API の api.UseMiddleware() で適用
//
// # ヘルパー
//
// GetClaims はハンドラ内で Context から認証済み Claims を取得します。
// 未認証（Claims 不在）の場合は ErrUnauthenticated を返します。
//
// # アダプター
//
// Chi → Huma 変換が必要な場合は、アプリケーション層で humachi パッケージを使用してください。
// gloudia は特定の Huma アダプター（humachi 等）に依存しないため、
// 変換はアプリケーション側で実装する必要があります。
//
// 変換例（アプリケーション側のコード）:
//
//	import "github.com/danielgtaylor/huma/v2/humachi"
//
//	// Chi → Huma: Chi ミドルウェアを Huma API に適用する
//	func ChiToHuma(chiMw func(http.Handler) http.Handler) func(huma.Context, func(huma.Context)) {
//	    return func(ctx huma.Context, next func(huma.Context)) {
//	        w := ctx.BodyWriter().(http.ResponseWriter)
//	        r := humachi.GetRequest(ctx)
//	        chiMw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
//	            next(ctx)
//	        })).ServeHTTP(w, r)
//	    }
//	}
package middleware
