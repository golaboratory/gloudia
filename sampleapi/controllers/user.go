package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	controller "github.com/golaboratory/gloudia/api/controllers"
)

type User struct {
	controller.BaseController
	ControllerName string
}

type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func (c *User) RegisterRoutes(api huma.API, basePath string) {

	c.Api = api
	c.BasePath = basePath
	c.BaseController.ControllerName = c.ControllerName

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodGet,
			Path:        "/{name}",
			Summary:     "Get a greeting",
			Description: "Get a greeting for a person by name.",
			HandlerFunc: c.GetGreeting,
		}), c.GetGreeting)
}

func (c *User) GetGreeting(_ context.Context, input *struct {
	Name string `path:"name" maxLength:"30" example:"world" doc:"Name to greet"`
}) (*GreetingOutput, error) {
	resp := &GreetingOutput{}
	resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)
	return resp, nil
}
