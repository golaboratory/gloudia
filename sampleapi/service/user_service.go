package service

import (
	"net/http"

	controller "github.com/golaboratory/gloudia/api/controllers"
	"github.com/golaboratory/gloudia/api/middleware"
	"github.com/golaboratory/gloudia/api/service"
	model "github.com/golaboratory/gloudia/sampleapi/structure/user"
)

type User struct {
	service.BaseService
}

func (u *User) ValidateForLogin(input *model.LoginInput) bool {
	if input == nil {
		u.AddInvalid("input", "Input is required")
		return false
	}

	if input.Body.UserId == "" {
		return false
	}

	if input.Body.Password == "" {
		return false
	}

	return u.IsValid()
}

func (u *User) TryLogin(input *model.LoginInput) (*controller.Res[model.AuthorizationInfo], error) {

	resp := &controller.Res[model.AuthorizationInfo]{}
	payload := model.AuthorizationInfo{}
	resp.Body.SummaryMessage = "Login failed"
	resp.Body.HasInvalidParams = true

	token, err := middleware.CreateJWT(middleware.Claims{UserID: "1", Role: "admin"})
	if err != nil {
		return nil, err
	}

	payload.Token = token
	payload.ID = 1
	payload.Username = "admin"

	resp.Body.Payload = payload
	resp.SetCookie = http.Cookie{
		Name:     "Authorization",
		Value:    token,
		HttpOnly: true,
		//Secure:   c.ApiConfig.EnableSSL,
	}

	return resp, nil
}
