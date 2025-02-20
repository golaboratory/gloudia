package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golaboratory/gloudia/core/text"
	"github.com/golang-jwt/jwt"
)

var (
	JWTMiddlewareName = "JWTAuthMiddleware"
)

// JWTMiddleware は、Huma用のミドルウェアとして、JWT認証を行う処理を実装します。
// 引数 secret は JWT の署名検証用シークレットです。
// Authorization ヘッダから "Bearer " トークンを抽出し、トークンの署名と有効性をチェックします。
func JWTMiddleware(api huma.API, secret string) func(ctx huma.Context, next func(huma.Context)) {

	return func(ctx huma.Context, next func(huma.Context)) {

		// 操作に認証が必要かどうかを確認
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			if _, ok := opScheme[JWTMiddlewareName]; ok {
				isAuthorizationRequired = true
				break
			}
		}

		// 認証が不要な場合は次のハンドラを実行
		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		// Authorization ヘッダを取得
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}
		// Bearer トークンかどうかを確認
		if !strings.HasPrefix(authHeader, "Bearer ") {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// JWT トークンを解析および検証
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// HMAC 署名方法のみ許容
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("予期しない署名方法: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// クレーム情報を取得し、コンテキストに保存
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			authInfo := claims["auth"].(string)
			ctx = huma.WithValue(ctx, "auth-info", authInfo)
		} else {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// 認証成功: 次のハンドラを実行
		next(ctx)
	}
}

// CreateJWT は、指定されたユーザー情報を含む JWT を生成します。
// 引数 secret は JWT の署名検証用シークレットです。
// 引数 claims は JWT に含めるクレーム情報です。
func CreateJWT(secret string, claims jwt.MapClaims, exp time.Duration, authInfo Claims) (string, error) {

	claims["exp"] = time.Now().Add(exp).Unix()

	auth, err := text.SerializeJson[Claims](authInfo)
	if err != nil {
		auth = ""
	}
	claims["auth"] = auth

	// 新しいトークンを生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// トークンに署名
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
