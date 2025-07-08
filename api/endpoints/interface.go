package endpoints

import (
	"github.com/danielgtaylor/huma/v2"
)

// Endpoint はAPIエンドポイントを表すインターフェースです。
// RegisterRoutes メソッドでエンドポイントのルーティングを登録します。
type Endpoint interface {
	// RegisterRoutes はエンドポイントのルートをAPIに登録します。
	RegisterRoutes(huma.API)
}
