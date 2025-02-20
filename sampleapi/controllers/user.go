package controllers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	controller "github.com/golaboratory/gloudia/api/controllers"
)

type User struct {
	controller.BaseController
}

type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func (c *User) RegisterRoutes(api huma.API, basePath string) {

	c.Api = api
	c.BasePath = basePath

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodGet,
			Path:        "{id}",
			Summary:     "Find User Entity By Id",
			Description: "ユーザマスタのIDを条件に、エンティティ情報を取得する",
			HandlerFunc: c.FindById,
			Controller:  c,
		}),
		c.FindById)

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodPost,
			Path:        "",
			Summary:     "Create User Entity",
			Description: "ユーザマスタのエンティティ情報を登録する",
			HandlerFunc: c.Create,
			Controller:  c,
		}),
		c.Create)

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodPut,
			Path:        "{id}",
			Summary:     "Update User Entity By Id",
			Description: "ユーザマスタのIDを条件に、エンティティ情報を更新する",
			HandlerFunc: c.Update,
			Controller:  c,
		}),
		c.Update)

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodDelete,
			Path:        "{id}",
			Summary:     "Delete User Entity By Id",
			Description: "ユーザマスタのIDを条件に、エンティティ情報を削除する",
			HandlerFunc: c.Delete,
			Controller:  c,
		}),
		c.Delete)

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodGet,
			Path:        "",
			Summary:     "Find User Entity List",
			Description: "ユーザマスタのエンティティ情報を取得する",
			HandlerFunc: c.GetAll,
			Controller:  c,
		}),
		c.GetAll)

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodGet,
			Path:        "",
			Summary:     "Find User Entity List With Delete Flag",
			Description: "ユーザマスタのエンティティ情報を取得する（削除フラグ有り）",
			HandlerFunc: c.GetAllWithDeleted,
			Controller:  c,
		}),
		c.GetAllWithDeleted)

}

func (c *User) FindById(_ context.Context, input *controller.PathIdParam) (*struct{}, error) {
	return nil, nil
}

func (c *User) Create(_ context.Context, input *struct{}) (*struct{}, error) {
	return nil, nil
}
func (c *User) Update(_ context.Context, input *controller.PathIdParam) (*struct{}, error) {
	return nil, nil
}
func (c *User) Delete(_ context.Context, input *controller.PathIdParam) (*struct{}, error) {
	return nil, nil
}
func (c *User) GetAll(_ context.Context, input *struct{}) (*struct{}, error) {
	return nil, nil
}
func (c *User) GetAllWithDeleted(_ context.Context, input *struct{}) (*struct{}, error) {
	return nil, nil
}
