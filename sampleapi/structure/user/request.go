package user

type LoginInput struct {
	Body struct {
		UserId   string `json:"userId" example:"user1" doc:"User ID"`
		Password string `json:"password" example:"password" doc:"Password"`
	}
}
