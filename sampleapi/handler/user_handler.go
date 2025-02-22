package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	controller "github.com/golaboratory/gloudia/api/controllers"
	sec "github.com/golaboratory/gloudia/core/security"
	"github.com/golaboratory/gloudia/sampleapi/service"
	structure "github.com/golaboratory/gloudia/sampleapi/structure/user"
)

type User struct {
	controller.BaseController
}

func (c *User) RegisterRoutes(api huma.API) {

	c.Api = api
	c.LoadConfig()

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodGet,
			Path:        "/{id}",
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
			Path:        "/{id}",
			Summary:     "Update User Entity By Id",
			Description: "ユーザマスタのIDを条件に、エンティティ情報を更新する",
			HandlerFunc: c.Update,
			Controller:  c,
		}),
		c.Update)

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:      http.MethodDelete,
			Path:        "/{id}",
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

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:         http.MethodPost,
			Path:           "/login",
			AllowAnonymous: true,
			Summary:        "Try Login",
			Description:    "ログインを試行する",
			HandlerFunc:    c.TryLogin,
			Controller:     c,
		}),
		c.TryLogin)

	huma.Register(api,
		c.CreateOperation(controller.OperationParams{
			Method:         http.MethodGet,
			Path:           "/enc/{text}",
			AllowAnonymous: true,
			Summary:        "Encript Text",
			Description:    "文字列を暗号化する",
			HandlerFunc:    c.EncriptText,
			Controller:     c,
		}),
		c.EncriptText)

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

func (c *User) TryLogin(ctx context.Context, input *structure.LoginInput) (*controller.Res[structure.AuthorizationInfo], error) {

	model := service.User{}
	model.Context = &ctx

	ok, msg, err := model.ValidateForLogin(input)
	if err != nil {
		return nil, err
	}
	if !ok {
		return controller.ResponseInvalid[structure.AuthorizationInfo](
			msg,
			model.InvalidList,
		)
	}

	resp, cookie, err := model.TryLogin(input)
	if err != nil {
		return nil, err
	}

	res, err := controller.ResponseOk[structure.AuthorizationInfo](*resp, "")
	res.SetCookie = cookie

	return res, err

}

func (c *User) EncriptText(ctx context.Context, input *controller.PathTextParam) (*controller.Res[controller.ResEncryptedText], error) {
	crypt := sec.Cryptography{}
	enc := crypt.HashString(input.Text)
	result := &controller.ResEncryptedText{
		EncryptedText: enc,
	}

	return controller.ResponseOk[controller.ResEncryptedText](*result, "")
}
