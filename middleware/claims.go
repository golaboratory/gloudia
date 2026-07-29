package middleware

import (
	"context"

	"github.com/golaboratory/gloudia/auth"
	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrUnauthenticated はコンテキストに認証情報が存在しない場合のエラーです。
	ErrUnauthenticated = ergo.NewSentinel("unauthenticated: claims not found in context")
)

// GetClaims はコンテキストから認証済みの Claims を取得します。
// 認証ミドルウェアによって設定された Claims が存在しない場合はエラーを返します。
// 使用例:
//
//	claims, err := middleware.GetClaims(ctx.Context())
//	if err != nil {
//	    return err
//	}
func GetClaims(ctx context.Context) (*auth.Claims, error) {
	claims, ok := ctx.Value(KeyClaims).(*auth.Claims)
	if !ok || claims == nil {
		return nil, ErrUnauthenticated
	}
	return claims, nil
}
