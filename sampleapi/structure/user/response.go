package user

type AuthorizationInfo struct {
	Token    string `json:"token" doc:"トークン"`
	UserId   int64  `json:"userId" doc:"ユーザID"`
	UserName string `json:"userName" doc:"ユーザ名"`
}
