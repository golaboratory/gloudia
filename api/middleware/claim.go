package middleware

type Claims struct {
	AuthKey string `json:"auth_key"`
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
}
