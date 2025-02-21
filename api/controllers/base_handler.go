package controller

import (
	"fmt"
	"strings"

	apiConfig "github.com/golaboratory/gloudia/api/config"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golaboratory/gloudia/api/middleware"
	"github.com/golaboratory/gloudia/core/ref"
	"github.com/golaboratory/gloudia/core/text"
)

// BaseController は基本的なコントローラを表現する構造体です。
// APIとの連携や、コントローラ名、基本パスを管理します。
type BaseController struct {
	Api       huma.API
	ApiConfig apiConfig.ApiConfig
}

// OperationParams は、API操作作成時に使用するパラメータ情報を保持する構造体です。
// メソッド、パス、概要、説明、ハンドラ関数などの情報が含まれます。
type OperationParams struct {
	Method         string
	Path           string
	Summary        string
	Description    string
	AllowAnonymous bool
	HandlerFunc    any
	Controller     any
}

func (c *BaseController) LoadConfig() {
	c.ApiConfig = apiConfig.ApiConfig{}
	if err := c.ApiConfig.Load(); err != nil {
		fmt.Println("Error: ", err)
	}
}

// CreateOperation は、指定されたパラメータからAPI操作（Operation）を生成します。
// 各操作に一意のOperationIDを割り当て、タグ情報を設定します。
func (c *BaseController) CreateOperation(param OperationParams) huma.Operation {

	var tags []string
	conf := apiConfig.ApiConfig{}
	if err := conf.Load(); err != nil {
		fmt.Println("Error: ", err)
	}

	controllerName, err := ref.GetStructName(param.Controller)
	if err != nil {
		fmt.Println("Error: ", err)
		controllerName = "unknown"
		tags = append(tags, "_has-error")
	} else {
		tags = append(tags, controllerName)
	}

	controllerName = strings.ToLower(controllerName)

	operationId, _ := ref.GetFuncName(param.HandlerFunc)
	operationId = controllerName + "-" + text.ConvertCamelToKebab(operationId)

	path := c.ApiConfig.RootPath + "/" + controllerName + param.Path

	if strings.Contains(path, "//") {
		tags = append(tags, "_has-error")
	}

	// tags = append(tags, "controller_"+controllerName)
	// tags = append(tags, "method_"+param.Method)

	security := []map[string][]string{}
	if !param.AllowAnonymous && conf.EnableJWT {
		security = []map[string][]string{
			{middleware.JWTMiddlewareName: {}},
		}
	}

	fmt.Println("--------------------")
	fmt.Println("Controller Name: ", controllerName)
	fmt.Println("Operation ID: ", operationId)
	fmt.Println("Method: ", param.Method)
	fmt.Println("Path: ", path)
	fmt.Println("Summary: ", param.Summary)
	fmt.Println("AllowAnonymous: ", param.AllowAnonymous)
	fmt.Println("")

	return huma.Operation{
		OperationID: operationId,
		Method:      param.Method,
		Path:        path,
		Summary:     param.Summary,
		Description: param.Description,
		Tags:        tags,
		Security:    security,
	}
}
