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

type Binder struct {
	APITitle    string
	APIVersion  string
	RootPath    string
	JWTValidate func(middleware.Claims) (bool, error)
}

func (b *Binder) Bind(endpoints []Endpoint) (humacli.CLI, error) {

	conf, err := config.New[apiConfig.ApiConfig]()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	cli := humacli.New(func(hooks humacli.Hooks, _ *struct{}) {
		// Create a new router & API
		router := chi.NewMux()

		if conf.EnableStatic {

			sconf, err := config.New[apiConfig.StaticConfig]()
			if err != nil {
				fmt.Println("Error: ", err)
			}

			// Serve static files
			fileServer := http.FileServer(http.Dir(sconf.HostingDirectory))
			router.Get(fmt.Sprintf("%s/*", sconf.BindingPath),
				func(w http.ResponseWriter, r *http.Request) {
					http.StripPrefix(fmt.Sprintf("%s/", sconf.BindingPath), fileServer).ServeHTTP(w, r)
				},
			)
		}

		if conf.EnableSpaProxy {

			pconf, err := config.New[apiConfig.ProxyConfig]()
			if err != nil {
				fmt.Println("Error: ", err)
			}

			targetURL, err := url.Parse(pconf.BackendURL)
			if err != nil {
				fmt.Printf("リバースプロキシURLの解析に失敗: %v\n", err)
			} else {
				proxy := httputil.NewSingleHostReverseProxy(targetURL)
				router.Get(fmt.Sprintf("%s/*", pconf.BindingPath), func(w http.ResponseWriter, r *http.Request) {
					r.URL.Scheme = targetURL.Scheme
					r.URL.Host = targetURL.Host
					r.Host = targetURL.Host
					r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
					r.Header.Set("X-Origin-Host", targetURL.Host)

					if strings.HasPrefix(r.URL.Path, pconf.BindingPath) {
						fmt.Println("Path: ", r.URL.Path)
						r.URL.Path = r.URL.Path[len(pconf.BindingPath):]
					}

					proxy.ServeHTTP(w, r)
				})
			}
		}

		config := huma.DefaultConfig(b.APITitle, b.APIVersion)

		if conf.EnableJWT {
			config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
				middleware.JWTMiddlewareName: {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			}
		}

		api := humachi.New(router, config)

		if conf.EnableJWT {
			// Add JWT middleware
			api.UseMiddleware(middleware.JWTMiddleware(api, b.JWTValidate))
		}

		// Register all endpoints
		for _, endpoint := range endpoints {
			endpoint.RegisterRoutes(api)
		}

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
