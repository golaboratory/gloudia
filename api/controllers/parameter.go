package controller

import (
	"mime/multipart"
	"net/http"

	"github.com/golaboratory/gloudia/api/service"
)

// PathIdParam は、エンティティの識別子 (ID) をパスパラメータとして受け取るための構造体です。
type PathIdParam struct {
	Id int64 `path:"id" example:"1" doc:"ID of the entity"`
}

// PathTextParam は、エンティティのテキストをパスパラメータとして受け取るための構造体です。
type PathTextParam struct {
	Text string `path:"text" example:"text" doc:"encripte text"`
}

// ReqFiles は、リクエストで送信されるファイル群を保持するための構造体です。
type ReqFiles struct {
	RawBody multipart.Form
}

// ResEncryptedText は、暗号化されたテキストをレスポンスとして返す構造体です。
type ResEncryptedText struct {
	EncryptedText string `json:"encryptedText" example:"encrypted text" doc:"Encrypted text"`
}

// ResponseBody は、APIレスポンスの本体を表す汎用的な構造体です。
// フィールド:
//   - SummaryMessage: レスポンスの概要メッセージ。
//   - HasInvalidParams: パラメータの妥当性チェック結果。
//   - InvalidParamList: 不正なパラメータのリスト（サービスから提供）。
//   - Payload: 任意のレスポンスデータ。
type ResponseBody[T any] struct {
	SummaryMessage           string `json:"summaryMessage" example:"Invalid parameters" doc:"Summary message"`
	HasInvalidParams         bool   `json:"hasInvalidParams" example:"false" doc:"Invalid parameters flag"`
	service.InvalidParamList `json:"invalidParamList" doc:"List of invalid parameters"`
	Payload                  T `json:"payload" doc:"Response payload"`
}

// Res は、HTTPレスポンスを表す汎用的な構造体です。
// フィールド:
//   - SetCookie: 設定されるCookie情報。
//   - Body: レスポンス本体（ResponseBody形式）。
type Res[T any] struct {
	SetCookie http.Cookie     `header:"Set-Cookie"`
	Body      ResponseBody[T] `json:"body" doc:"Response body"`
}

// BinalyResponse は、バイナリデータを含むHTTPレスポンスを表す構造体です。
// フィールド:
//   - ContentType: レスポンスのContent-Typeヘッダー。
//   - Body: バイナリ形式のデータ。
type BinalyResponse struct {
	ContentType string `header:"Content-Type"`
	Body        []byte
}
