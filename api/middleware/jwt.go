package middleware

import (
	"fmt"
	apiConfig "github.com/golaboratory/gloudia/api/config"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golaboratory/gloudia/core/text"
	gjwt "github.com/golang-jwt/jwt"
)

// JWTMiddlewareName はJWT認証ミドルウェアの名前を示します。
var JWTMiddlewareName = "JWTAuthMiddleware"

// JWTSecret はJWTの署名検証用シークレットを保持します。
var JWTSecret = "BHqQTg99LmSk$Q,_xe*LM+!P*5PKnR~n"

// JWTMiddleware はJWT認証ミドルウェアです。
// APIリクエストに対してAuthorizationヘッダ内のJWTトークンを検証し、認証情報をコンテキストに設定します。
func JWTMiddleware(api huma.API, secret string) func(ctx huma.Context, next func(huma.Context)) {

	if secret != "" {
		JWTSecret = secret
	}

	conf := apiConfig.ApiConfig{}
	if err := conf.Load(); err != nil {
		fmt.Println("Error: ", err)
	}

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
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized1")
			return
		}
		// Bearer トークンかどうかを確認
		if !strings.HasPrefix(authHeader, "Bearer ") {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized2")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if conf.EnableCookieToken {

			// Cookieからトークンを取得
			var authCookie string
			if c, err := huma.ReadCookie(ctx, "Authorization"); err != nil {
				authCookie = ""
			} else {
				authCookie = c.Value
			}

			if authCookie != "" {
				tokenString = authCookie
			}

		}

		// JWTトークンの解析および検証
		token, err := gjwt.Parse(tokenString, func(token *gjwt.Token) (interface{}, error) {
			// HMAC署名方式であることを確認
			if _, ok := token.Method.(*gjwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(JWTSecret), nil
		})
		if err != nil || !token.Valid {
			fmt.Printf("JWTMiddleware: %s\n", err)
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized3")
			return
		}

		// クレーム情報を取得し、コンテキストに保存
		if claims, ok := token.Claims.(gjwt.MapClaims); ok && token.Valid {
			authInfo := claims["auth"].(string)
			ctx = huma.WithValue(ctx, "auth-info", authInfo)
		} else {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized4")
			return
		}

		// 認証成功: 次のハンドラを実行
		next(ctx)
	}
}

// CreateJWT は、指定されたユーザー情報を含むJWTトークンを生成します。
// 引数secretは署名検証用シークレット、expはトークンの有効期限、authInfoは認証情報です。
func CreateJWT(secret string, exp time.Duration, authInfo Claims) (string, error) {

	if secret != "" {
		JWTSecret = secret
	}

	claims := gjwt.MapClaims{}
	claims["exp"] = time.Now().Add(exp).Unix()

	auth, err := text.SerializeJson[Claims](authInfo)
	if err != nil {
		auth = ""
	}
	claims["auth"] = auth

	// HS256方式で新しいトークンを生成
	token := gjwt.NewWithClaims(gjwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
