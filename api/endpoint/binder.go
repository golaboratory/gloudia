package endpoints

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8888"`
}

type Binder struct {
	APITitle   string
	APIVersion string
	RootPath   string
}

func (b *Binder) Bind(endpoints []Endpoint) (humacli.CLI, error) {

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new router & API
		router := chi.NewMux()

		//router.Use(jwtauth.Verifier(tokenAuth))
		//router.Use(logger.New())

		// Serve static files
		fileServer := http.FileServer(http.Dir("./static/"))
		router.Get("/app/*",
			func(w http.ResponseWriter, r *http.Request) {
				http.StripPrefix("/app/", fileServer).ServeHTTP(w, r)
			},
		)

		api := humachi.New(router, huma.DefaultConfig(b.APITitle, b.APIVersion))

		// Register all endpoints
		for _, endpoint := range endpoints {
			endpoint.RegisterRoutes(api, b.RootPath)
		}

		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			err := http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
			if err != nil {
				fmt.Println("Error starting server:", err)
			}
		})
	})

	return cli, nil
}
