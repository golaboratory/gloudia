package user

import (
	"github.com/golaboratory/gloudia/sampleapi/repository/db"
)

type AuthorizationInfo struct {
	Token string `json:"token" doc:"トークン"`
	db.MUser
}
