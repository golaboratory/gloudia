package service

import (
	"context"
	"fmt"

	apiconfig "github.com/golaboratory/gloudia/api/config"
	"github.com/golaboratory/gloudia/core/config"
)

// InvalidParam は、バリデーションエラーが発生したパラメータの情報を保持する構造体です。
// フィールド:
//   - Name: パラメータ名
//   - Message: エラーメッセージ
type InvalidParam struct {
	Name    string `json:"name" example:"id" doc:"Parameter name"`
	Message string `json:"message" example:"Id is required" doc:"Error message"`
}

// InvalidParamList は、バリデーションエラーが発生したパラメータのリストを表現する型です。
type InvalidParamList []InvalidParam

// BaseService は、サービスの基本機能を提供する構造体です。
// コンテキストの管理やバリデーションエラーの管理機能を提供します。
type BaseService struct {
	Context     *context.Context
	InvalidList InvalidParamList
	APIConfig   apiconfig.ApiConfig
}

func (b *BaseService) LoadConfig() {

	conf, err := config.New[apiconfig.ApiConfig]()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	b.APIConfig = conf

}

// AddInvalid は、バリデーションエラーのリストに新しいエラー情報を追加します。
//
// パラメータ:
//   - name: エラーが発生したパラメータ名
//   - message: エラーメッセージ
func (b *BaseService) AddInvalid(name, message string) {
	b.InvalidList = append(
		b.InvalidList,
		InvalidParam{Name: name, Message: message})
}

// IsValid は、バリデーションエラーが存在しないかを確認します。
//
// 戻り値:
//   - bool: エラーが存在しない場合は true、存在する場合は false
func (b *BaseService) IsValid() bool {
	return len(b.InvalidList) == 0
}

// ClearInvalid は、バリデーションエラーのリストをクリアします。
func (b *BaseService) ClearInvalid() {
	b.InvalidList = []InvalidParam{}
}
