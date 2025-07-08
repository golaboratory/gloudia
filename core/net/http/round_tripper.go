package http

import (
	"net/http"
	"time"
)

// retryStatus はリトライ対象となるHTTPステータスコードのマップです。
var retryStatus = map[int]struct{}{
	http.StatusInternalServerError: {},
	http.StatusTooManyRequests:     {},
}

// GloudiaRoundTripper はHTTPリクエストのリトライ処理を行うカスタムRoundTripperです。
//   - t: 元となるhttp.RoundTripper
//   - maxRetry: 最大リトライ回数
//   - wait: リトライ間の待機時間
type GloudiaRoundTripper struct {
	t        http.RoundTripper
	maxRetry int
	wait     time.Duration
}

// RoundTrip はHTTPリクエストを送信し、リトライ対象のステータスコードの場合はリトライを行います。
// 引数:
//   - req: 送信するHTTPリクエスト
//
// 戻り値:
//   - *http.Response: レスポンス
//   - error: エラー情報
func (r GloudiaRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	var res *http.Response
	var err error

	for c := 0; c < r.maxRetry; c++ {
		res, err = r.t.RoundTrip(req)

		if err != nil {
			return nil, req.Context().Err()
		}

		var status = http.StatusRequestTimeout
		if res != nil {
			status = res.StatusCode
		}

		// リトライ対象外のステータスコードの場合は終了
		if _, ok := retryStatus[status]; !ok {
			return res, err
		}

		select {
		case <-req.Context().Done(): // キャンセルされた場合は終了
			return nil, req.Context().Err()
		case <-time.After(r.wait): // 待機時間を挟んでリトライ
		}
	}

	return res, err
}
