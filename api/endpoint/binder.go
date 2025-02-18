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
	RootPath string
}

func (b *Binder) Bind(endpoints []Endpoint) (humacli.CLI, error) {

	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Create a new router & API
		router := chi.NewMux()
		//router.Use(jwtauth.Verifier(tokenAuth))
		//router.Use(logger.New())
		api := humachi.New(router, huma.DefaultConfig("My API", "1.0.0"))

		for _, endpoint := range endpoints {
			endpoint.RegisterRoutes(&api, b.RootPath)
		}

		hooks.OnStart(func() {
			fmt.Printf("Starting server on port %d...\n", options.Port)
			http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
		})
	})

	return cli, nil
}
