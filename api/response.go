package api

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// internalServerErrorMessage は 500 応答でクライアントへ返す汎用メッセージです。
// 内部エラーの詳細（SQL 断片・ファイルパス・接続情報等）を露出させないために使用します。
const internalServerErrorMessage = "サーバー内部エラーが発生しました"

// FieldName はバリデーションエラーが発生したフィールドの名前を表す型です。
type FieldName string

// ErrorMessage はエラーの内容を表すメッセージの型です。
type ErrorMessage string

// InvalidItem はバリデーションエラーの詳細（フィールドごとのエラー）を表します。
// フロントエンドのフォームバリデーション（赤枠表示など）に使用されます。
type InvalidItem map[FieldName]ErrorMessage

// unifiedResponder は UnifiedResponseBody を埋め込んだ構造体が自動的に満たす非公開インターフェースです。
// ジェネリックヘルパー関数の型制約として使用されます。
// getUnifiedBody がポインタレシーバであるため、埋め込んだ構造体 T ではなく
// そのポインタ *T のみがこの制約を満たします（&T{...} を渡してください）。
type unifiedResponder interface {
	getUnifiedBody() *UnifiedResponseBody
}

// UnifiedResponseBody はレスポンスのボディ部分を定義する構造体です。
// ペイロード（DTOなど）は含みません。埋め込み構造体として使用されることを想定しています。
// エラー情報は huma.ErrorModel を通じて返却されるため、このボディには成功時の情報のみを保持します。
type UnifiedResponseBody struct {
	// SummaryMessage はトースト通知などに表示するための要約メッセージです。
	// 成功時は「保存しました」などが入ります。
	SummaryMessage string `json:"summaryMessage"`
}

// getUnifiedBody は unifiedResponder インターフェースの実装です。
// UnifiedResponseBody を埋め込んだ構造体はこのメソッドを自動的に継承します。
// ポインタレシーバのため、埋め込んだ構造体はポインタ型のみが unifiedResponder を満たします。
func (b *UnifiedResponseBody) getUnifiedBody() *UnifiedResponseBody {
	return b
}

// setSummary はレスポンスに要約メッセージを設定します。
// UnifiedResponseBody がポインタ埋め込み（*UnifiedResponseBody）かつ nil の場合に
// nil ポインタ参照で panic しないよう、nil チェックを行います（値埋め込み推奨）。
func setSummary[T unifiedResponder](resp T, message string) {
	if body := resp.getUnifiedBody(); body != nil {
		body.SummaryMessage = message
	}
}

// SetSuccess はレスポンスの要約メッセージ（SummaryMessage）に message を設定し、
// humaハンドラーの戻り値としてそのまま return できる (T, nil) のペアを返却する
// ジェネリックヘルパー関数です。error は常に nil です。
//
// 使用例:
//
//	return api.SetSuccess(resp, "保存しました")
func SetSuccess[T unifiedResponder](resp T, message string) (T, error) {
	setSummary(resp, message)
	return resp, nil
}

// SetInvalid はレスポンスにバリデーションエラーを設定し、humaハンドラーの戻り値として
// そのまま return できる (T, error) のペアを返却するジェネリックヘルパー関数です。
// resp には要約メッセージのみが設定され、フィールドごとの詳細は戻り値の error 側で運ばれます。
// error には huma.ErrorModel (HTTP 422) が格納され、InvalidItem の各フィールドエラーは
// huma.ErrorDetail に変換されます。InvalidItem はマップのため、ErrorDetail の順序は不定です。
//
// 使用例:
//
//	return api.SetInvalid(resp, "入力エラー", details)
func SetInvalid[T unifiedResponder](resp T, message string, details InvalidItem) (T, error) {
	setSummary(resp, message)

	errs := make([]error, 0, len(details))
	for field, msg := range details {
		errs = append(errs, &huma.ErrorDetail{
			Message:  string(msg),
			Location: "body." + string(field),
		})
	}

	return resp, huma.NewError(http.StatusUnprocessableEntity, message, errs...)
}

// SetUnauthorized はレスポンスに認証エラーを設定し、humaハンドラーの戻り値として
// そのまま return できる (T, error) のペアを返却するジェネリックヘルパー関数です。
// error には huma.ErrorModel (HTTP 401) が格納されます。
// message が空文字列の場合、デフォルトで "認証エラー" が使用されます。
//
// 使用例:
//
//	return api.SetUnauthorized(resp, "ログインが必要です")
func SetUnauthorized[T unifiedResponder](resp T, message string) (T, error) {
	if message == "" {
		message = "認証エラー"
	}
	setSummary(resp, message)
	return resp, huma.NewError(http.StatusUnauthorized, message)
}

// SetBadRequest はレスポンスに不正リクエストエラーを設定し、humaハンドラーの戻り値として
// そのまま return できる (T, error) のペアを返却するジェネリックヘルパー関数です。
// error には huma.ErrorModel (HTTP 400) が格納されます。
// message が空文字列の場合、デフォルトで "不正なリクエストです" が使用されます。
//
// 使用例:
//
//	return api.SetBadRequest(resp, "テナントIDが不正です")
func SetBadRequest[T unifiedResponder](resp T, message string) (T, error) {
	if message == "" {
		message = "不正なリクエストです"
	}
	setSummary(resp, message)
	return resp, huma.NewError(http.StatusBadRequest, message)
}

// SetError はレスポンスに予期しない内部エラーを設定し、humaハンドラーの戻り値として
// そのまま return できる (T, error) のペアを返却するジェネリックヘルパー関数です。
// err の内容は slog.Error でサーバーログにのみ記録され、クライアントには一切返却されません。
// 要約メッセージ（設定済みの SummaryMessage は上書き）と huma.ErrorModel (HTTP 500) の Detail は
// いずれも固定の内部エラーメッセージに置き換えられます。err が nil の場合は error も nil を返却します。
//
// 使用例:
//
//	return api.SetError(resp, err)
func SetError[T unifiedResponder](resp T, err error) (T, error) {
	if err != nil {
		// 内部エラーの詳細はクライアントに漏らさず、サーバーログにのみ記録する。
		// クライアントには汎用メッセージのみを返す。
		slog.Error("internal server error", slog.String("error", err.Error()))
		setSummary(resp, internalServerErrorMessage)
		return resp, huma.NewError(http.StatusInternalServerError, internalServerErrorMessage)
	}
	return resp, nil
}

// SetForbidden はレスポンスにアクセス権限エラーを設定し、humaハンドラーの戻り値として
// そのまま return できる (T, error) のペアを返却するジェネリックヘルパー関数です。
// error には huma.ErrorModel (HTTP 403) が格納されます。
// message が空文字列の場合、デフォルトで "アクセス権限がありません" が使用されます。
//
// 使用例:
//
//	return api.SetForbidden(resp, "このリソースへのアクセス権限がありません")
func SetForbidden[T unifiedResponder](resp T, message string) (T, error) {
	if message == "" {
		message = "アクセス権限がありません"
	}
	setSummary(resp, message)
	return resp, huma.NewError(http.StatusForbidden, message)
}

// SetNotFound はレスポンスに対象未検出エラーを設定し、humaハンドラーの戻り値として
// そのまま return できる (T, error) のペアを返却するジェネリックヘルパー関数です。
// error には huma.ErrorModel (HTTP 404) が格納されます。
// message が空文字列の場合、デフォルトで "対象が見つかりません" が使用されます。
//
// 使用例:
//
//	return api.SetNotFound(resp, "指定された予約が見つかりません")
func SetNotFound[T unifiedResponder](resp T, message string) (T, error) {
	if message == "" {
		message = "対象が見つかりません"
	}
	setSummary(resp, message)
	return resp, huma.NewError(http.StatusNotFound, message)
}
