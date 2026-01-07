package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: 実際のAPIを叩くテストはコストと認証情報が必要なため、
// ここでは構造体の初期化やパラメータ変換ロジックを中心にテストします。
// 統合テストを行う場合は、環境変数からAPIキーを読み込むか、モックを使用してください。

func TestNewClient(t *testing.T) {
	cfg := Config{
		APIKey: "dummy_key",
		OrgID:  "dummy_org",
	}
	client := NewClient(cfg)

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

// TestCreateChatCompletion_Mock は本来であれば go-openai のモックインターフェースを使うべきですが、
// sasha/go-openai は struct ベースのクライアントであり、インターフェース化されていない部分があるため、
// 厳密なユニットテストには工夫が必要です。
// ここでは、ライブラリのラッパーとしての責務（パラメータマッピングなど）はシンプルであるため、
// 初期化テストに留め、ロジックが複雑化した場合にモック導入を検討します。
