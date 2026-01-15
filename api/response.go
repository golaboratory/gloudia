package api

// FieldName はバリデーションエラーが発生したフィールドの名前を表す型です。
type FieldName string

// ErrorMessage はエラーの内容を表すメッセージの型です。
type ErrorMessage string

// InvalidItem はバリデーションエラーの詳細（フィールドごとのエラー）を表します。
// フロントエンドのフォームバリデーション（赤枠表示など）に使用されます。
type InvalidItem map[FieldName]ErrorMessage

// IUnifiedResponse はレスポンスのステータス設定を行うインターフェースです。
type IUnifiedResponse interface {
	SetSuccess(message string)
	SetInvalid(message string, details InvalidItem)
	SetError(err error)
}

// UnifiedResponseBody はレスポンスのボディ部分を定義する構造体です。
// ペイロード（DTOなど）は含みません。埋め込み構造体として使用されることを想定しています。
type UnifiedResponseBody struct {
	// IsInvalid はバリデーションエラーやビジネスロジックエラーがある場合に true となります。
	IsInvalid bool `json:"isInvalid"`

	// InvalidList は IsInvalid=true の場合に、具体的なフィールドごとのエラー情報を格納します。
	// エラーがない場合は omitempty により JSON に含まれません。
	InvalidList InvalidItem `json:"errors,omitempty"`

	// SummaryMessage はトースト通知などに表示するための要約メッセージです。
	// 成功時は「保存しました」、エラー時は「入力内容を確認してください」などが入ります。
	SummaryMessage string `json:"summaryMessage"`

	// Error はビジネスロジックエラーの詳細を表します。
	// エラーがない場合は omitempty により JSON に含まれません。
	Error error `json:"error,omitempty"`
}

// SetSuccess は成功時のステータスをセットします。
func (b *UnifiedResponseBody) SetSuccess(message string) {
	b.IsInvalid = false
	b.SummaryMessage = message
}

// SetInvalid はバリデーションエラー時のステータスをセットします。
func (b *UnifiedResponseBody) SetInvalid(message string, details InvalidItem) {
	b.IsInvalid = true
	b.SummaryMessage = message
	b.InvalidList = details
}

// SetError はビジネスロジックエラー時のステータスをセットします。
func (b *UnifiedResponseBody) SetError(err error) {
	b.IsInvalid = true
	if err != nil {
		b.SummaryMessage = err.Error()
		b.Error = err
	}
}
