package controller

import (
	"mime/multipart"
	"net/http"

	"github.com/golaboratory/gloudia/api/service"
)

// PathIdParam はエンティティの識別子（ID）をパスパラメータとして受け取るための構造体です。
type PathIdParam struct {
	Id int64 `path:"id" example:"1" doc:"エンティティのID"`
}

// PathTextParam はエンティティのテキストをパスパラメータとして受け取るための構造体です。
type PathTextParam struct {
	Text string `path:"text" example:"text" doc:"暗号化テキスト"`
}

// ReqFiles はリクエストで送信されるファイル群を保持するための構造体です。
type ReqFiles struct {
	RawBody multipart.Form
}

// ResEncryptedText は暗号化されたテキストをレスポンスとして返す構造体です。
type ResEncryptedText struct {
	EncryptedText string `json:"encryptedText" example:"encrypted text" doc:"暗号化されたテキスト"`
}

// ResponseBody はAPIレスポンスの本体を表す汎用的な構造体です。
//   - SummaryMessage: レスポンスの概要メッセージ
//   - HasInvalidParams: パラメータの妥当性チェック結果
//   - InvalidParamList: 不正なパラメータのリスト（サービスから提供）
//   - Payload: 任意のレスポンスデータ
type ResponseBody[T any] struct {
	SummaryMessage           string `json:"summaryMessage" example:"Invalid parameters" doc:"概要メッセージ"`
	HasInvalidParams         bool   `json:"hasInvalidParams" example:"false" doc:"不正パラメータ有無フラグ"`
	service.InvalidParamList `json:"invalidParamList" doc:"不正なパラメータのリスト"`
	Payload                  T `json:"payload" doc:"レスポンスペイロード"`
}

// Res はHTTPレスポンスを表す汎用的な構造体です。
//   - SetCookie: 設定されるCookie情報
//   - Body: レスポンス本体（ResponseBody形式）
type Res[T any] struct {
	SetCookie http.Cookie     `header:"Set-Cookie"`
	Body      ResponseBody[T] `json:"body" doc:"レスポンス本体"`
}

// BinalyResponse はバイナリデータを含むHTTPレスポンスを表す構造体です。
//   - ContentType: レスポンスのContent-Typeヘッダー
//   - Body: バイナリ形式のデータ
type BinalyResponse struct {
	ContentType string `header:"Content-Type"`
	Body        []byte
}
