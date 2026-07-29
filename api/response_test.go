package api

import (
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResponse is a struct for testing UnifiedResponseBody embedding.
type TestResponse struct {
	UnifiedResponseBody
	Payload map[string]string `json:"payload,omitempty"`
}

// ptrEmbedResponse は *UnifiedResponseBody をポインタ埋め込みした（誤用）構造体です。
type ptrEmbedResponse struct {
	*UnifiedResponseBody
	Payload map[string]string `json:"payload,omitempty"`
}

// TestSetters_NilPointerEmbeddedBody は、ポインタ埋め込みされた *UnifiedResponseBody が
// nil の場合でもヘルパーが panic しないこと（nil ガード）を検証します。
func TestSetters_NilPointerEmbeddedBody(t *testing.T) {
	assert.NotPanics(t, func() {
		resp := &ptrEmbedResponse{} // 埋め込みポインタは nil
		_, _ = SetSuccess(resp, "ok")
		_, _ = SetBadRequest(resp, "bad")
		_, _ = SetError(resp, huma.Error400BadRequest("dummy"))
		_, _ = SetNotFound(resp, "")
	})
}

func TestSetSuccess(t *testing.T) {
	payload := map[string]string{"foo": "bar"}
	msg := "Operation successful"

	resp := &TestResponse{
		Payload: payload,
	}
	got, err := SetSuccess(resp, msg)

	require.NoError(t, err)
	assert.Same(t, resp, got)
	assert.Equal(t, msg, got.SummaryMessage)
	assert.Equal(t, payload, got.Payload)
}

func TestSetInvalid(t *testing.T) {
	msg := "Validation failed"
	details := InvalidItem{
		"field1": "required",
	}

	resp := &TestResponse{}
	got, err := SetInvalid(resp, msg, details)

	assert.Same(t, resp, got)
	require.Error(t, err)
	var errModel huma.StatusError
	require.ErrorAs(t, err, &errModel)
	assert.Equal(t, http.StatusUnprocessableEntity, errModel.GetStatus())
	assert.Equal(t, msg, got.SummaryMessage)
}

func TestSetInvalid_MultipleDetails(t *testing.T) {
	msg := "Validation failed"
	details := InvalidItem{
		"name":  "required",
		"email": "invalid format",
	}

	resp := &TestResponse{}
	got, err := SetInvalid(resp, msg, details)

	assert.Same(t, resp, got)
	require.Error(t, err)
	errModel, ok := err.(*huma.ErrorModel)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, errModel.Status)
	assert.Len(t, errModel.Errors, 2)
}

func TestSetInvalid_EmptyDetails(t *testing.T) {
	msg := "Validation failed"
	details := InvalidItem{}

	resp := &TestResponse{}
	got, err := SetInvalid(resp, msg, details)

	assert.Same(t, resp, got)
	require.Error(t, err)
	errModel, ok := err.(*huma.ErrorModel)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, errModel.Status)
	assert.Empty(t, errModel.Errors)
}

func TestSetUnauthorized(t *testing.T) {
	t.Run("カスタムメッセージ", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetUnauthorized(resp, "ログインが必要です")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusUnauthorized, errModel.GetStatus())
		assert.Equal(t, "ログインが必要です", got.SummaryMessage)
	})

	t.Run("空メッセージでデフォルト値を使用", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetUnauthorized(resp, "")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusUnauthorized, errModel.GetStatus())
		assert.Equal(t, "認証エラー", got.SummaryMessage)
	})
}

func TestSetUnauthorized_ReturnDirectly(t *testing.T) {
	handler := func() (*TestResponse, error) {
		resp := &TestResponse{}
		return SetUnauthorized(resp, "セッションが無効です")
	}

	got, err := handler()
	require.Error(t, err)
	assert.Equal(t, "セッションが無効です", got.SummaryMessage)
	var errModel huma.StatusError
	require.ErrorAs(t, err, &errModel)
	assert.Equal(t, http.StatusUnauthorized, errModel.GetStatus())
}

func TestSetError(t *testing.T) {
	t.Run("With Error", func(t *testing.T) {
		origErr := huma.Error500InternalServerError("business logic error: secret table=users")

		resp := &TestResponse{}
		got, err := SetError(resp, origErr)

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusInternalServerError, errModel.GetStatus())
		// 内部エラーの詳細をクライアントに漏らさず、汎用メッセージのみ返すこと。
		assert.Equal(t, internalServerErrorMessage, got.SummaryMessage)
		assert.NotContains(t, errModel.Error(), "secret table=users")
	})

	t.Run("With Nil", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetError(resp, nil)

		assert.Same(t, resp, got)
		require.NoError(t, err)
		assert.Empty(t, got.SummaryMessage)
	})
}

func TestSetSuccess_ReturnDirectly(t *testing.T) {
	// humaハンドラーで return api.SetSuccess(resp, msg) と記述できることを検証
	handler := func() (*TestResponse, error) {
		resp := &TestResponse{Payload: map[string]string{"id": "123"}}
		return SetSuccess(resp, "作成しました")
	}

	got, err := handler()
	require.NoError(t, err)
	assert.Equal(t, "作成しました", got.SummaryMessage)
	assert.Equal(t, "123", got.Payload["id"])
}

func TestSetInvalid_ReturnDirectly(t *testing.T) {
	handler := func() (*TestResponse, error) {
		resp := &TestResponse{}
		return SetInvalid(resp, "入力エラー", InvalidItem{"name": "必須です"})
	}

	got, err := handler()
	require.Error(t, err)
	assert.Equal(t, "入力エラー", got.SummaryMessage)
	var errModel huma.StatusError
	require.ErrorAs(t, err, &errModel)
	assert.Equal(t, http.StatusUnprocessableEntity, errModel.GetStatus())
}

func TestSetError_ReturnDirectly(t *testing.T) {
	handler := func() (*TestResponse, error) {
		resp := &TestResponse{}
		return SetError(resp, huma.Error500InternalServerError("DB接続エラー"))
	}

	got, err := handler()
	require.Error(t, err)
	assert.NotEmpty(t, got.SummaryMessage)
	var errModel huma.StatusError
	require.ErrorAs(t, err, &errModel)
	assert.Equal(t, http.StatusInternalServerError, errModel.GetStatus())
}

// ===========================================================================
// 以下、リファクタリング保護テスト
//
// `SetForbidden` (HTTP 403) / `SetNotFound` (HTTP 404) も同じ規約に従う。
// 既存ヘルパーの振る舞いを以下の観点で固定し、新規ヘルパー追加時の整合性を担保する：
//   1. 全ヘルパーが `huma.StatusError` を返却すること（`huma.NewError` ベース）
//   2. 期待するステータスコードを返却すること
//   3. `SummaryMessage` がレスポンスボディに正しく設定されること
//   4. デフォルトメッセージ機構（空文字 → デフォルト値）が一貫していること
// ===========================================================================

// TestSetBadRequest はこれまで未テストだった SetBadRequest の振る舞いを固定する。
// SetForbidden / SetNotFound 追加時に同じシグネチャで実装されることの基準ケースとなる。
func TestSetBadRequest(t *testing.T) {
	t.Run("カスタムメッセージ", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetBadRequest(resp, "テナントIDが不正です")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusBadRequest, errModel.GetStatus())
		assert.Equal(t, "テナントIDが不正です", got.SummaryMessage)
	})

	t.Run("空メッセージでデフォルト値を使用", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetBadRequest(resp, "")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusBadRequest, errModel.GetStatus())
		assert.Equal(t, "不正なリクエストです", got.SummaryMessage)
	})
}

// TestSetBadRequest_ReturnDirectly は huma ハンドラからの直接 return パターンを保護する。
func TestSetBadRequest_ReturnDirectly(t *testing.T) {
	handler := func() (*TestResponse, error) {
		resp := &TestResponse{}
		return SetBadRequest(resp, "入力値が不正です")
	}

	got, err := handler()
	require.Error(t, err)
	assert.Equal(t, "入力値が不正です", got.SummaryMessage)
	var errModel huma.StatusError
	require.ErrorAs(t, err, &errModel)
	assert.Equal(t, http.StatusBadRequest, errModel.GetStatus())
}

// TestResponseHelpers_StatusCodeMatrix は全エラーヘルパーが期待ステータスを返すことを
// テーブル駆動で検証する。新規ヘルパー追加時にこの表へ行を追加することで、
// 「ステータスコード → ヘルパー関数」のマッピングが網羅されていることを保証する。
func TestResponseHelpers_StatusCodeMatrix(t *testing.T) {
	tests := []struct {
		name       string
		invoke     func(resp *TestResponse) (*TestResponse, error)
		wantStatus int
		wantMsg    string
	}{
		{
			name: "SetUnauthorized → 401",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetUnauthorized(resp, "認証エラー")
			},
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "認証エラー",
		},
		{
			name: "SetBadRequest → 400",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetBadRequest(resp, "不正リクエスト")
			},
			wantStatus: http.StatusBadRequest,
			wantMsg:    "不正リクエスト",
		},
		{
			name: "SetForbidden → 403",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetForbidden(resp, "アクセス権限がありません")
			},
			wantStatus: http.StatusForbidden,
			wantMsg:    "アクセス権限がありません",
		},
		{
			name: "SetNotFound → 404",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetNotFound(resp, "対象が見つかりません")
			},
			wantStatus: http.StatusNotFound,
			wantMsg:    "対象が見つかりません",
		},
		{
			name: "SetInvalid → 422",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetInvalid(resp, "入力エラー", InvalidItem{"f": "msg"})
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantMsg:    "入力エラー",
		},
		{
			name: "SetError → 500",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetError(resp, huma.Error500InternalServerError("内部エラー"))
			},
			wantStatus: http.StatusInternalServerError,
			// SetError は内部エラーの詳細を漏らさず汎用メッセージのみ返す。
			wantMsg: internalServerErrorMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &TestResponse{}
			got, err := tt.invoke(resp)

			require.Error(t, err)
			assert.Same(t, resp, got, "ヘルパーは引数の resp ポインタをそのまま返すべき")

			var errModel huma.StatusError
			require.ErrorAs(t, err, &errModel, "返却エラーは huma.StatusError を満たすべき")
			assert.Equal(t, tt.wantStatus, errModel.GetStatus(), "ステータスコードが期待値と一致すべき")
			assert.Equal(t, tt.wantMsg, got.SummaryMessage, "SummaryMessage が引数メッセージで上書きされるべき")
		})
	}
}

// TestResponseHelpers_DefaultMessageContract は「空文字メッセージ → デフォルト値設定」
// の契約を全該当ヘルパーで一貫していることを保護する。
func TestResponseHelpers_DefaultMessageContract(t *testing.T) {
	tests := []struct {
		name        string
		invoke      func(resp *TestResponse) (*TestResponse, error)
		wantDefault string
	}{
		{
			name: "SetUnauthorized 空メッセージ → デフォルト",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetUnauthorized(resp, "")
			},
			wantDefault: "認証エラー",
		},
		{
			name: "SetBadRequest 空メッセージ → デフォルト",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetBadRequest(resp, "")
			},
			wantDefault: "不正なリクエストです",
		},
		{
			name: "SetForbidden 空メッセージ → デフォルト",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetForbidden(resp, "")
			},
			wantDefault: "アクセス権限がありません",
		},
		{
			name: "SetNotFound 空メッセージ → デフォルト",
			invoke: func(resp *TestResponse) (*TestResponse, error) {
				return SetNotFound(resp, "")
			},
			wantDefault: "対象が見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &TestResponse{}
			got, _ := tt.invoke(resp)
			assert.Equal(t, tt.wantDefault, got.SummaryMessage,
				"空メッセージ時はヘルパー固有のデフォルト日本語メッセージが入るべき")
		})
	}
}

// TestSetInvalid_LocationPrefix は SetInvalid が ErrorDetail.Location に
// "body." プレフィックスを付与する仕様を保護する。
// フロントエンドのフォームエラー表示（フィールド単位ハイライト）はこの仕様に依存。
//
// リファクタリング保護:
//   - SetInvalid のロケーション仕様を変更すると、利用側フロントエンドの
//     エラー表示が崩れる。本テストで仕様を凍結する。
func TestSetInvalid_LocationPrefix(t *testing.T) {
	resp := &TestResponse{}
	_, err := SetInvalid(resp, "入力エラー", InvalidItem{
		"applicantName": "必須です",
		"email":         "形式が不正です",
	})

	require.Error(t, err)
	errModel, ok := err.(*huma.ErrorModel)
	require.True(t, ok)
	require.Len(t, errModel.Errors, 2)

	for _, detail := range errModel.Errors {
		require.NotNil(t, detail, "Errors の各要素は nil ではないべき")
		assert.Contains(t, detail.Location, "body.",
			"Location には 'body.' プレフィックスが付与されるべき (実値: %s)", detail.Location)
	}
}

// TestSetError_RedactsInternalError は SetError が内部エラーの詳細を
// クライアントに漏らさず、汎用メッセージのみを返すことを検証する（情報漏洩対策）。
func TestSetError_RedactsInternalError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		secret string
	}{
		{
			name:   "huma エラー",
			err:    huma.Error500InternalServerError("DB接続失敗 host=db.internal:5432"),
			secret: "db.internal:5432",
		},
		{
			name:   "標準 error",
			err:    huma.NewError(http.StatusInternalServerError, "pq: relation \"users\" does not exist"),
			secret: "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &TestResponse{}
			got, err := SetError(resp, tt.err)
			require.Error(t, err)
			// 汎用メッセージのみが返り、内部詳細は SummaryMessage にもエラーにも含まれない。
			assert.Equal(t, internalServerErrorMessage, got.SummaryMessage)
			assert.NotContains(t, got.SummaryMessage, tt.secret)
			var errModel huma.StatusError
			require.ErrorAs(t, err, &errModel)
			assert.NotContains(t, errModel.Error(), tt.secret)
		})
	}
}

// TestSetForbidden は SetForbidden の振る舞いを検証する。
func TestSetForbidden(t *testing.T) {
	t.Run("カスタムメッセージ", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetForbidden(resp, "このリソースへのアクセス権限がありません")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusForbidden, errModel.GetStatus())
		assert.Equal(t, "このリソースへのアクセス権限がありません", got.SummaryMessage)
	})

	t.Run("空メッセージでデフォルト値を使用", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetForbidden(resp, "")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusForbidden, errModel.GetStatus())
		assert.Equal(t, "アクセス権限がありません", got.SummaryMessage)
	})
}

// TestSetForbidden_ReturnDirectly は huma ハンドラからの直接 return パターンを保護する。
func TestSetForbidden_ReturnDirectly(t *testing.T) {
	handler := func() (*TestResponse, error) {
		resp := &TestResponse{}
		return SetForbidden(resp, "他人の予約は参照できません")
	}

	got, err := handler()
	require.Error(t, err)
	assert.Equal(t, "他人の予約は参照できません", got.SummaryMessage)
	var errModel huma.StatusError
	require.ErrorAs(t, err, &errModel)
	assert.Equal(t, http.StatusForbidden, errModel.GetStatus())
}

// TestSetNotFound は SetNotFound の振る舞いを検証する。
func TestSetNotFound(t *testing.T) {
	t.Run("カスタムメッセージ", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetNotFound(resp, "指定された予約が見つかりません")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusNotFound, errModel.GetStatus())
		assert.Equal(t, "指定された予約が見つかりません", got.SummaryMessage)
	})

	t.Run("空メッセージでデフォルト値を使用", func(t *testing.T) {
		resp := &TestResponse{}
		got, err := SetNotFound(resp, "")

		assert.Same(t, resp, got)
		require.Error(t, err)
		var errModel huma.StatusError
		require.ErrorAs(t, err, &errModel)
		assert.Equal(t, http.StatusNotFound, errModel.GetStatus())
		assert.Equal(t, "対象が見つかりません", got.SummaryMessage)
	})
}

// TestSetNotFound_ReturnDirectly は huma ハンドラからの直接 return パターンを保護する。
func TestSetNotFound_ReturnDirectly(t *testing.T) {
	handler := func() (*TestResponse, error) {
		resp := &TestResponse{}
		return SetNotFound(resp, "ユーザーが存在しません")
	}

	got, err := handler()
	require.Error(t, err)
	assert.Equal(t, "ユーザーが存在しません", got.SummaryMessage)
	var errModel huma.StatusError
	require.ErrorAs(t, err, &errModel)
	assert.Equal(t, http.StatusNotFound, errModel.GetStatus())
}

// TestUnifiedResponseBody_EmbedContract は UnifiedResponseBody の埋め込み構造体が
// unifiedResponder インターフェースを自動的に満たす仕様を保護する。
// 利用側で新しいレスポンス型を定義する際、この仕様が前提となる。
func TestUnifiedResponseBody_EmbedContract(t *testing.T) {
	// TestResponse は UnifiedResponseBody を埋め込んでいるため、
	// getUnifiedBody() を自動的に持つ → ジェネリックヘルパーの型制約を満たす
	resp := &TestResponse{Payload: map[string]string{"key": "value"}}

	// SetSuccess にそのまま渡せることを確認
	got, err := SetSuccess(resp, "ok")
	require.NoError(t, err)
	assert.Equal(t, "ok", got.SummaryMessage)
	assert.Equal(t, "value", got.Payload["key"], "Payload は SetSuccess 呼び出しで破壊されないべき")
}
