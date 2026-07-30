package api

import (
	"github.com/danielgtaylor/huma/v2"
)

// Handler はAPIハンドラーが実装すべき共通インターフェースです。
// ルーティングの登録処理を統一的に扱うために使用されます。
type Handler interface {
	// RegisterRoutes は指定されたAPIインスタンスに対してルート定義を登録します。
	// middlewares はこのハンドラーグループに適用するミドルウェア、
	// SecurityScheme はルートに適用する OpenAPI セキュリティスキーム名（空文字列は指定なし）、
	// rootPath はベースパスを指定します。
	RegisterRoutes(api huma.API, middlewares huma.Middlewares, SecurityScheme string, rootPath string)
}
