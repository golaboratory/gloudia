package main

import (
	endpoints "github.com/golaboratory/gloudia/api/endpoint"
	"github.com/golaboratory/gloudia/sampleapi/handler"
)

var (
	Endpoints = []endpoints.Endpoint{
		&handler.User{},
	}
)

func main() {
	binder := &endpoints.Binder{
		RootPath:   "/api",
		APITitle:   "Sample API",
		APIVersion: "1.0.0",
	}
	cli, err := binder.Bind(Endpoints)

	if err != nil {
		panic(err)
	}

	cli.Run()

}
