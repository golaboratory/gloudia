package http

import (
	"net/http"
	"time"
)

var retryStatus = map[int]struct{}{
	http.StatusInternalServerError: {},
	http.StatusTooManyRequests:     {},
}

type GloudiaRoundTripper struct {
	t        http.RoundTripper
	maxRetry int
	wait     time.Duration
}

func (r GloudiaRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {

	var res *http.Response
	var err error

	for c := 0; c < r.maxRetry; c++ {
		res, err = r.t.RoundTrip(req)

		var status = 0
		if res == nil {
			status = http.StatusRequestTimeout
		} else {
			status = res.StatusCode
		}

		// リトライすべきステータスか確認する
		if _, ok := retryStatus[status]; !ok {
			return res, err
		}

		select {
		case <-req.Context().Done(): // contextの期限が来たら終了
			return nil, req.Context().Err()
		case <-time.After(r.wait): // リトライ間隔を待つ
		}
	}

	return res, err
}
