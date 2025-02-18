package controller

import (
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
	Tags        []string
	HandlerFunc any
}

func (c *BaseController) CreateOperation(param OperationParams) huma.Operation {

	operationId, _ := ref.GetFuncName(param.HandlerFunc)
	operationId = c.ControllerName + "-" + operationId

	path := c.BasePath + "/" + c.ControllerName + param.Path

	param.Tags = append(param.Tags, c.ControllerName)
	param.Tags = append(param.Tags, param.Method)

	return huma.Operation{
		OperationID: operationId,
		Method:      param.Method,
		Path:        path,
		Summary:     param.Summary,
		Description: param.Description,
		Tags:        param.Tags,
	}
}
