package middleware

// Claims はJWTなどの認証情報を格納する構造体です。
//   - AuthKey: 認証キー
//   - UserID: ユーザーID
//   - Role: ユーザーのロール
type Claims struct {
	AuthKey string `json:"auth_key"`
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
}
