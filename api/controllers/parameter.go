package controller

// PathIdParam は、エンティティの識別子 (ID) をパスパラメータとして受け取るための構造体です。
// フィールド:
//   - Id: エンティティのID。例として 1 が指定されます。
type PathIdParam struct {
	Id int64 `path:"id" example:"1" doc:"ID of the entity"`
}

type Res[T any] struct {
	Body struct {
		SummaryMessage   string `json:"summaryMessage" example:"Invalid parameters" doc:"Summary message"`
		HasInvalidParams bool   `json:"hasInvalidParams" example:"false" doc:"Invalid parameters flag"`
		InvalidParamList []struct {
			Name    string `json:"name" example:"id" doc:"Parameter name"`
			Message string `json:"message" example:"Id is required" doc:"Error message"`
		} `json:"invalidParamList" doc:"List of invalid parameters"`
		Payload T `json:"payload" doc:"Response payload"`
	}
}
