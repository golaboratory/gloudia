package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"

	"github.com/golaboratory/gloudia/api/middleware"
	"github.com/golaboratory/gloudia/api/service"

	pg "github.com/golaboratory/gloudia/core/db"
	"github.com/golaboratory/gloudia/core/security"
	rep "github.com/golaboratory/gloudia/sampleapi/repository/db"
	model "github.com/golaboratory/gloudia/sampleapi/structure/user"
)

type User struct {
	service.BaseService
}

func (u *User) ValidateForLogin(input *model.LoginInput) (bool, string, error) {
	if input == nil {
		u.AddInvalid("userId", "Input is required")
		u.AddInvalid("password", "Input is required")
		return false, "", nil
	}

	if input.Body.UserId == "" {
		u.AddInvalid("userId", "Input is required")
	}

	if input.Body.Password == "" {
		u.AddInvalid("password", "Input is required")
	}

	if !u.IsValid() {
		return false, "", nil
	}

	conn, _ := pg.NewPostgresConnection()
	defer func(conn *pgx.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}(conn, context.Background())
	q := rep.New(conn)

	cript := security.Cryptography{}
	hash := cript.HashString(input.Body.Password)
	fmt.Println(hash)
	user, err := q.TryLogin(
		*u.Context,
		rep.TryLoginParams{
			LoginID:      input.Body.UserId,
			PasswordHash: hash,
		})

	if err != nil {
		fmt.Println(err)
	}

	if user.ID == 0 {
		u.AddInvalid("userId", "Input is required2")
		u.AddInvalid("password", "Input is required")
		return false, "", nil
	}

	return u.IsValid(), "", nil
}

func (u *User) TryLogin(input *model.LoginInput) (*model.AuthorizationInfo, http.Cookie, error) {

	u.LoadConfig()

	payload := model.AuthorizationInfo{}

	conn, _ := pg.NewPostgresConnection()
	defer func(conn *pgx.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}(conn, context.Background())
	q := rep.New(conn)

	cript := security.Cryptography{}
	hash := cript.HashString(input.Body.Password)

	user, err := q.TryLogin(
		*u.Context,
		rep.TryLoginParams{
			LoginID:      input.Body.UserId,
			PasswordHash: hash,
		})
	if err != nil {
		return nil, http.Cookie{}, err
	}

	token, err := middleware.CreateJWT(middleware.Claims{UserID: strconv.FormatInt(user.ID, 10), Role: "admin"})
	if err != nil {
		return nil, http.Cookie{}, err
	}
	payload.MUser = user
	payload.Token = token

	return &payload,
		http.Cookie{
			Name:     "Authorization",
			Value:    token,
			HttpOnly: true,
			Secure:   u.APIConfig.EnableSSL,
		}, nil
}
