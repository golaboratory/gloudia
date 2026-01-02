package api

// FieldName はバリデーションエラーが発生したフィールドの名前を表す型です。
type FieldName string

// ErrorMessage はエラーの内容を表すメッセージの型です。
type ErrorMessage string

// InvalidItem はバリデーションエラーの詳細（フィールドごとのエラー）を表します。
// フロントエンドのフォームバリデーション（赤枠表示など）に使用されます。
type InvalidItem map[FieldName]ErrorMessage

// UnifiedResponseBody はレスポンスのボディ部分を定義する構造体です。
// [T any] は成功時のペイロード（DTOなど）の型を指定します。
// 正常系、エラー系（バリデーションエラー含む）を問わず、常にこの形式で返却します。
type UnifiedResponseBody[T any] struct {
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

	// Payload は API のメインとなるデータです。
	// エラー時などデータがない場合は omitempty により JSON に含まれません。
	Payload T `json:"payload,omitempty"`
}

// UnifiedResponse はシステム共通のAPIレスポンス形式です。
// JSONレスポンスのルートとして使用されます。
type UnifiedResponse[T any] struct {
	Body UnifiedResponseBody[T] `json:",inline"`
}

// NewSuccessResponse は正常系のレスポンスを生成するヘルパー関数です。
func NewSuccessResponse[T any](payload T, message string) *UnifiedResponse[T] {

	return &UnifiedResponse[T]{
		Body: UnifiedResponseBody[T]{
			IsInvalid:      false,
			SummaryMessage: message,
			Payload:        payload,
		},
	}
}

// NewInvalidResponse はエラー系（バリデーションエラー等）のレスポンスを生成するヘルパー関数です。
func NewInvalidResponse[T any](message string, details InvalidItem) *UnifiedResponse[T] {
	return &UnifiedResponse[T]{
		Body: UnifiedResponseBody[T]{
			IsInvalid:      true,
			SummaryMessage: message,
			InvalidList:    details,
		},
	}
}

// NewErrorResponse はエラー系（バリデーションエラー等）のレスポンスを生成するヘルパー関数です。
func NewErrorResponse[T any](err error) *UnifiedResponse[T] {
	return &UnifiedResponse[T]{
		Body: UnifiedResponseBody[T]{
			IsInvalid:      true,
			SummaryMessage: err.Error(),
			Error:          err,
		},
	}
}
