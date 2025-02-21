package service

import (
	"net/http"

	"github.com/golaboratory/gloudia/api/middleware"
	"github.com/golaboratory/gloudia/api/service"
	model "github.com/golaboratory/gloudia/sampleapi/structure/user"
)

type User struct {
	service.BaseService
}

func (u *User) ValidateForLogin(input *model.LoginInput) (bool, string) {
	if input == nil {
		u.AddInvalid("userId", "Input is required")
		u.AddInvalid("password", "Input is required")
		return false, ""
	}

	if input.Body.UserId == "" {
		u.AddInvalid("userId", "Input is required")
	}

	if input.Body.Password == "" {
		u.AddInvalid("password", "Input is required")
	}

	return u.IsValid(), ""
}

func (u *User) TryLogin(input *model.LoginInput) (*model.AuthorizationInfo, http.Cookie, error) {

	u.LoadConfig()

	payload := model.AuthorizationInfo{}

	token, err := middleware.CreateJWT(middleware.Claims{UserID: "1", Role: "admin"})
	if err != nil {
		return nil, http.Cookie{}, err
	}

	payload.Token = token
	payload.ID = 1
	payload.Username = "admin"

	return &payload,
		http.Cookie{
			Name:     "Authorization",
			Value:    token,
			HttpOnly: true,
			Secure:   u.APIConfig.EnableSSL,
		}, nil
}
