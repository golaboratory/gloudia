package controller

import (
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golaboratory/gloudia/core/ref"
)

type BaseController struct {
	ControllerName string
	Api            huma.API
	BasePath       string
}

type OperationParams struct {
	Method      string
	Path        string
	Summary     string
	Description string
	HandlerFunc any
}

func (c *BaseController) CreateOperation(param OperationParams) huma.Operation {

	operationId, _ := ref.GetFuncName(param.HandlerFunc)
	operationId = c.ControllerName + "-" + operationId

	path := c.BasePath + "/" + c.ControllerName + param.Path

	var tags []string

	tags = append(tags, "controller_"+c.ControllerName)
	tags = append(tags, "method_"+param.Method)

	fmt.Println("--------------------")
	fmt.Println("Controller Name: ", c.ControllerName)
	fmt.Println("Operation ID: ", operationId)
	fmt.Println("Method: ", param.Method)
	fmt.Println("Path: ", path)
	fmt.Println("Summary: ", param.Summary)

	return huma.Operation{
		OperationID: operationId,
		Method:      param.Method,
		Path:        path,
		Summary:     param.Summary,
		Description: param.Description,
		Tags:        tags,
	}
}
