package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	apiConfig "github.com/golaboratory/gloudia/api/config"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golaboratory/gloudia/core/config"
	"github.com/golaboratory/gloudia/core/text"
	gjwt "github.com/golang-jwt/jwt"
)

// JWTMiddlewareName はJWT認証ミドルウェアの名前を表します。
var JWTMiddlewareName = "JWTAuthMiddleware"

// JWTSecret はJWTの署名検証に使用するシークレットキーです。
var JWTSecret = "BHqQTg99LmSk$Q,_xe*LM+!P*5PKnR~n"

// JWTMiddleware はAPIリクエストに含まれるJWTトークンの検証を行い、
// 認証情報をコンテキストにセットするミドルウェアを返します。
// validate関数は認証情報(Claims)の妥当性を検証するために利用されます。
// また、環境設定に基づきCookieからトークンを取得する処理も行います。
func JWTMiddleware(api huma.API, validate func(Claims) (bool, error)) func(ctx huma.Context, next func(huma.Context)) {

	// APIサーバの設定情報を取得
	conf, err := config.New[apiConfig.ApiConfig]()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// 設定ファイルからJWTシークレットを取得
	if conf.JWTSecret != "" {
		JWTSecret = conf.JWTSecret
	}

	return func(ctx huma.Context, next func(huma.Context)) {

		// この操作が認証を必要とするか判定
		isAuthorizationRequired := false
		for _, opScheme := range ctx.Operation().Security {
			if _, ok := opScheme[JWTMiddlewareName]; ok {
				isAuthorizationRequired = true
				break
			}
		}

		// 認証不要の場合はそのまま次のハンドラへ
		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		// Authorizationヘッダからトークン文字列を取得
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			// ヘッダが無い場合は認証エラー
			err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized1")
			if err != nil {
				fmt.Printf("JWTMiddleware: %s\n", err)
			}
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Bearerスキームでない場合は認証エラー
			err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized2")
			if err != nil {
				fmt.Printf("JWTMiddleware: %s\n", err)
			}
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Cookieからトークンを取得する設定が有効な場合
		if conf.EnableCookieToken {
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

		// JWTトークンの解析および検証（署名方式はHMAC）
		token, err := gjwt.Parse(tokenString, func(token *gjwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*gjwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(JWTSecret), nil
		})
		if err != nil || !token.Valid {
			// トークンが無効な場合は認証エラー
			fmt.Printf("JWTMiddleware: %s\n", err)
			err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized3")
			if err != nil {
				fmt.Printf("JWTMiddleware: %s\n", err)
			}
			return
		}

		var authInfo string
		// クレームから認証情報を抽出
		if claims, ok := token.Claims.(gjwt.MapClaims); ok && token.Valid {
			authInfo = claims["auth"].(string)
		} else {
			// クレームが不正な場合は認証エラー
			err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized4")
			if err != nil {
				fmt.Printf("JWTMiddleware: %s\n", err)
			}
			return
		}

		if authInfo == "" {
			// 認証情報が空の場合は認証エラー
			err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized5")
			if err != nil {
				fmt.Printf("JWTMiddleware: %s\n", err)
			}
			return
		}

		var auth Claims
		// 認証情報をデシリアライズ
		if auth, err = text.DeserializeJson[Claims](authInfo); err != nil {
			err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized6")
			if err != nil {
				fmt.Printf("JWTMiddleware: %s\n", err)
			}
			return
		} else {
			// 認証情報の妥当性を検証
			if ok, err := validate(auth); !ok || err != nil {
				err := huma.WriteErr(api, ctx, http.StatusUnauthorized, "Unauthorized7")
				if err != nil {
					fmt.Printf("JWTMiddleware: %s\n", err)
				}
				return
			}
		}

		// 認証情報をコンテキストにセット
		ctx = huma.WithValue(ctx, "auth", auth)

		// 認証成功時に次のハンドラを実行
		next(ctx)
	}
}

// CreateJWT は認証情報を含むJWTトークンを生成します。
// 環境設定からトークンの有効期限およびシークレットキーを取得し、HS256方式で署名します。
func CreateJWT(authInfo Claims) (string, error) {

	conf, err := config.New[apiConfig.ApiConfig]()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	if conf.JWTSecret != "" {
		JWTSecret = conf.JWTSecret
	}

	claims := gjwt.MapClaims{}
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(conf.JWTExpireMinute)).Unix()

	auth, err := text.SerializeJson[Claims](authInfo)
	if err != nil {
		auth = ""
	}
	claims["auth"] = auth

	// HS256方式で新しいJWTトークンを生成
	token := gjwt.NewWithClaims(gjwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
