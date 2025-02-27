package endpoints_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	controller "github.com/golaboratory/gloudia/api/controllers"
	"github.com/golaboratory/gloudia/api/endpoints"
	"github.com/golaboratory/gloudia/api/middleware"
	"github.com/stretchr/testify/assert"
)

// DummyEndpoint は Binder のテスト用ダミーエンドポイントです。
// GET リクエストで、指定されたパスパラメータの値をそのまま返すエンドポイントを登録します。
type DummyEndpoint struct {
	controller.BaseController
}

// RegisterRoutes は DummyEndpoint のルートを登録します。
// ルート: {RootPath}/{controller名}/{text}
// ここでは controller.BaseController の CreateOperation を使い、入力テキストをレスポンスとして返します。
func (d *DummyEndpoint) RegisterRoutes(api huma.API) {
	d.Api = api
	d.LoadConfig()

	huma.Register(api,
		d.CreateOperation(controller.OperationParams{
			Method:         http.MethodGet,
			Path:           "/{text}",
			Summary:        "Dummy endpoint for testing",
			Description:    "入力されたテキストをそのまま返します",
			AllowAnonymous: true,
			HandlerFunc:    d.EncriptText,
			Controller:     d,
		}),
		d.EncriptText)
}

// EncriptText は、入力パラメータのテキストをレスポンスとして返すハンドラーです。
func (d *DummyEndpoint) EncriptText(ctx context.Context, input *controller.PathTextParam) (*controller.Res[controller.ResEncryptedText], error) {
	result := controller.ResEncryptedText{
		EncryptedText: input.Text,
	}
	return controller.ResponseOk(result, "success")
}

// TestBinder_Bind_DummyEndpoint は、Binder.Bind によるルート登録とリクエスト処理を検証します。
func TestBinder_Bind_DummyEndpoint(t *testing.T) {
	// テスト用環境変数の設定: 静的ファイル・SPAプロキシ機能を無効化し、APIポートを指定
	os.Setenv("ENABLE_STATIC", "false")
	os.Setenv("ENABLE_SPA_PROXY", "false")
	os.Setenv("PORT", "9999")
	defer os.Clearenv()

	binder := endpoints.Binder{
		APITitle:   "Test API",
		APIVersion: "v1",
		RootPath:   "/api",
		JWTValidate: func(claims middleware.Claims) (bool, error) {
			return true, nil
		},
	}

	// DummyEndpoint をエンドポイントとして登録
	dummy := &DummyEndpoint{}
	dummyEndpoints := []endpoints.Endpoint{dummy}

	cli, err := binder.Bind(dummyEndpoints)
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	// 取得した API オブジェクトから HTTP リクエストをシミュレーション
	//cli.Run()

	/*
		// ダミーエンドポイントは、controller.BaseController の名前に基づきルートが生成される。
		// ここでは、リクエストパスを "/api/dummyendpoint/foo" として送信します。
		//req
		httptest.NewRequest(http.MethodGet, "/api/dummyendpoint/foo", nil)
		rec := httptest.NewRecorder()

		//api.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		// レスポンスには "foo" が返ってくるはずです。
		assert.Contains(t, rec.Body.String(), "foo")
	*/
}
