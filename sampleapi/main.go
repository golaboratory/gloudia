package main

import (
	"fmt"

	apiConfig "github.com/golaboratory/gloudia/api/config"
	endpoints "github.com/golaboratory/gloudia/api/endpoint"
	"github.com/golaboratory/gloudia/api/middleware"
	"github.com/golaboratory/gloudia/sampleapi/handler"
)

// Endpoints は利用するエンドポイント（APIハンドラ）のリストです。
var (
	Endpoints = []endpoints.Endpoint{
		&handler.User{},
	}
)

func main() {

	// ① API設定のロード: 環境変数等から設定値を取得します。
	conf := apiConfig.ApiConfig{}
	if err := conf.Load(); err != nil {
		fmt.Println("Error: ", err)
	}

	// ② Binderの初期化: APIのルートパス、タイトル、バージョンを設定し、JWT検証関数を設定します。
	binder := &endpoints.Binder{
		RootPath:   conf.RootPath,   // APIのベースパスを設定
		APITitle:   conf.APITitle,   // APIのタイトルを設定
		APIVersion: conf.APIVersion, // APIのバージョンを設定
		JWTValidate: func(middleware.Claims) (bool, error) {
			// JWTトークンに含まれる認証情報の検証処理を行います。（サンプルでは常に認証成功）
			return true, nil
		},
	}

	// ③ エンドポイントのバインド: 定義したエンドポイントをBinderに登録します。
	cli, err := binder.Bind(Endpoints)
	if err != nil {
		// バインドエラーの場合はプログラムを停止します。
		panic(err)
	}

	// ④ CLIの起動: 構築したAPIサーバ（CLI）を起動します。
	cli.Run()
}
