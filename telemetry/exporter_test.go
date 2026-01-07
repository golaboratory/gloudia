package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOTLPExporter(t *testing.T) {
	// ダミーのOTLPレシーバー（コレクター）に見立てたHTTPサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// OTLPのリクエストを受け入れる
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// NewOTLPExporterのendpoint引数はスキーム(http://)なしのホスト:ポート形式を期待する
	endpoint := strings.TrimPrefix(server.URL, "http://")
	ctx := context.Background()

	t.Run("successfully creates exporter", func(t *testing.T) {
		// insecure=true で作成
		exporter, err := NewOTLPExporter(ctx, endpoint, true)

		assert.NoError(t, err)
		assert.NotNil(t, exporter)

		// 正常にシャットダウンできるか
		err = exporter.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("fails with invalid context", func(t *testing.T) {
		// キャンセル済みのコンテキストを渡す
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		exporter, err := NewOTLPExporter(cancelledCtx, endpoint, true)

		// otlptracehttp.New はコンテキストがキャンセルされているとエラーを返すことが多い
		assert.Error(t, err)
		assert.Nil(t, exporter)
	})
}
