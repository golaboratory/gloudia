package endpoints

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	apiConfig "github.com/golaboratory/gloudia/api/config"
	"github.com/golaboratory/gloudia/core/config"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"

	"github.com/golaboratory/gloudia/api/middleware"
)

// Binder はAPIエンドポイントのバインドやサーバ起動処理を管理する構造体です。
type Binder struct {
	APITitle    string
	APIVersion  string
	RootPath    string
	JWTValidate func(middleware.Claims) (bool, error)
}

// Bind はエンドポイント群をAPIサーバにバインドし、サーバを起動する関数です。
// 引数 endpoints にはバインドするエンドポイントのスライスを指定します。
// 戻り値はhumacli.CLIインスタンスとエラーです。
func (b *Binder) Bind(endpoints []Endpoint) (humacli.CLI, error) {

	// APIサーバの設定情報を取得
	conf, err := config.New[apiConfig.ApiConfig]()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	cli := humacli.New(func(hooks humacli.Hooks, _ *struct{}) {
		// ルーターとAPIの初期化
		router := chi.NewMux()

		// 静的ファイル配信の有効化
		if conf.EnableStatic {

			// 静的ファイル設定情報を取得
			staticConfig, err := config.New[apiConfig.StaticConfig]()
			if err != nil {
				fmt.Println("Error: ", err)
			}

			// 静的ファイルサーバの設定
			fileServer := http.FileServer(http.Dir(staticConfig.HostingDirectory))
			router.Get(fmt.Sprintf("%s/*", staticConfig.BindingPath),
				func(w http.ResponseWriter, r *http.Request) {
					// パスプレフィックスを除去して静的ファイルを配信
					http.StripPrefix(fmt.Sprintf("%s/", staticConfig.BindingPath), fileServer).ServeHTTP(w, r)
				},
			)
		}

		// SPAプロキシの有効化
		if conf.EnableSpaProxy {

			// プロキシ設定情報を取得
			proxyConfig, err := config.New[apiConfig.ProxyConfig]()
			if err != nil {
				fmt.Println("Error: ", err)
			}

			// バックエンドURLの解析
			targetURL, err := url.Parse(proxyConfig.BackendURL)
			if err != nil {
				fmt.Printf("リバースプロキシURLの解析に失敗: %v\n", err)
			} else {
				// リバースプロキシの設定
				proxy := httputil.NewSingleHostReverseProxy(targetURL)
				router.Get(fmt.Sprintf("%s/*", proxyConfig.BindingPath), func(w http.ResponseWriter, r *http.Request) {
					// リクエストのURLやヘッダーをプロキシ用に書き換え
					r.URL.Scheme = targetURL.Scheme
					r.URL.Host = targetURL.Host
					r.Host = targetURL.Host
					r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
					r.Header.Set("X-Origin-Host", targetURL.Host)

					// バインディングパスを除去
					if strings.HasPrefix(r.URL.Path, proxyConfig.BindingPath) {
						fmt.Println("Path: ", r.URL.Path)
						r.URL.Path = r.URL.Path[len(proxyConfig.BindingPath):]
					}

					// プロキシ経由でリクエストを転送
					proxy.ServeHTTP(w, r)
				})
			}
		}

		// humaのデフォルトAPI設定を作成
		defaultConfig := huma.DefaultConfig(b.APITitle, b.APIVersion)

		// JWT認証の有効化とセキュリティスキーマの設定
		if conf.EnableJWT {
			defaultConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
				middleware.JWTMiddlewareName: {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			}
		}

		// APIインスタンスの生成
		api := humachi.New(router, defaultConfig)

		// JWTミドルウェアの追加
		if conf.EnableJWT {
			api.UseMiddleware(middleware.JWTMiddleware(api, b.JWTValidate))
		}

		// すべてのエンドポイントを登録
		for _, endpoint := range endpoints {
			endpoint.RegisterRoutes(api)
		}

		// サーバ起動処理の登録
		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", conf.Port)
			err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router)
			if err != nil {
				fmt.Println("Error starting server:", err)
			}
		})
	})

	return cli, nil
}
