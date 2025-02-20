package endpoints

import (
	"fmt"
	apiConfig "github.com/golaboratory/gloudia/api/config"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"

	"github.com/golaboratory/gloudia/api/middleware"
)

type Binder struct {
	APITitle   string
	APIVersion string
	RootPath   string
}

func (b *Binder) Bind(endpoints []Endpoint) (humacli.CLI, error) {

	conf := apiConfig.ApiConfig{}
	if err := conf.Load(); err != nil {
		return nil, err
	}

	cli := humacli.New(func(hooks humacli.Hooks, _ *struct{}) {
		// Create a new router & API
		router := chi.NewMux()

		if conf.EnableStatic {
			// Serve static files
			fileServer := http.FileServer(http.Dir("./static/"))

			router.Get("/app/*",
				func(w http.ResponseWriter, r *http.Request) {
					http.StripPrefix("/app/", fileServer).ServeHTTP(w, r)
				},
			)
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
			api.UseMiddleware(
				middleware.JWTMiddleware(
					api,
					conf.JWTSecret))
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
