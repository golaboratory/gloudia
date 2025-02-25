package controller

import (
	"fmt"
	"strings"

	apiConfig "github.com/golaboratory/gloudia/api/config"
	"github.com/golaboratory/gloudia/api/service"

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

// ResponseInvalid は、無効なパラメータが検出された場合にエラーレスポンスを生成する関数です。
// message には概要メッセージ、invalidList には不正なパラメータのリストを指定します。
// 戻り値は、ResponseBody を含む Res[T] とエラーです。
func ResponseInvalid[T any](message string, invalidList service.InvalidParamList) (*Res[T], error) {

	result := &Res[T]{
		Body: ResponseBody[T]{
			SummaryMessage:   message,
			HasInvalidParams: true,
			InvalidParamList: invalidList,
		},
	}

	return result, nil
}

// ResponseOk は、正常なレスポンスを生成する関数です。
// payload にレスポンスデータ、message に概要メッセージを指定します。
// 戻り値は、ResponseBody を含む Res[T] とエラーです。
func ResponseOk[T any](payload T, message string) (*Res[T], error) {

	result := &Res[T]{
		Body: ResponseBody[T]{
			SummaryMessage:   message,
			HasInvalidParams: false,
			InvalidParamList: service.InvalidParamList{},
			Payload:          payload,
		},
	}

	return result, nil
}

// NewResponseBinary は、バイナリレスポンス用のhuma.Responseマップを生成する関数です。
// contentType にはレスポンスのContent-Type、description にはレスポンスの説明を指定します。
// 戻り値は、HTTPレスポンス定義のマップです。
func NewResponseBinary(contentType string, description string) map[string]*huma.Response {

	return map[string]*huma.Response{
		"200": {
			Description: description,
			Content: map[string]*huma.MediaType{
				contentType: {},
			},
		},
	}
}
