package controller

// PathIdParam は、エンティティの識別子 (ID) をパスパラメータとして受け取るための構造体です。
// フィールド:
//   - Id: エンティティのID。例として "1" が指定されます。
type PathIdParam struct {
	Id string `path:"id" example:"1" doc:"ID of the entity"`
}
