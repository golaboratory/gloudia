package controller

import (
	"net/http"

	"github.com/golaboratory/gloudia/api/service"
)

// PathIdParam は、エンティティの識別子 (ID) をパスパラメータとして受け取るための構造体です。
// フィールド:
//   - Id: エンティティのID。例として 1 が指定されます。
type PathIdParam struct {
	Id int64 `path:"id" example:"1" doc:"ID of the entity"`
}

type Res[T any] struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      struct {
		SummaryMessage           string `json:"summaryMessage" example:"Invalid parameters" doc:"Summary message"`
		HasInvalidParams         bool   `json:"hasInvalidParams" example:"false" doc:"Invalid parameters flag"`
		service.InvalidParamList `json:"invalidParamList" doc:"List of invalid parameters"`
		Payload                  T `json:"payload" doc:"Response payload"`
	}
}
