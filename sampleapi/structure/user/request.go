package user

type LoginInput struct {
	Body struct {
		UserId   string `json:"userId" example:"admin" doc:"User ID"`
		Password string `json:"password" example:"password" doc:"Password"`
	} `json:"body" doc:"Request body"`
}
