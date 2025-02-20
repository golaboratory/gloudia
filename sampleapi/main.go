package main

import (
	"fmt"
	apiConfig "github.com/golaboratory/gloudia/api/config"
	endpoints "github.com/golaboratory/gloudia/api/endpoint"
	"github.com/golaboratory/gloudia/sampleapi/handler"
)

var (
	Endpoints = []endpoints.Endpoint{
		&handler.User{},
	}
)

func main() {

	conf := apiConfig.ApiConfig{}
	if err := conf.Load(); err != nil {
		fmt.Println("Error: ", err)
	}

	binder := &endpoints.Binder{
		RootPath:   conf.RootPath,
		APITitle:   conf.APITitle,
		APIVersion: conf.APIVersion,
	}
	cli, err := binder.Bind(Endpoints)

	if err != nil {
		panic(err)
	}

	cli.Run()

}
