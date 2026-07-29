package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Do_NoRetryOn4xx(t *testing.T) {
	var accessCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&accessCount, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	conf := DefaultConfig()
	conf.MaxRetries = 3
	conf.RetryWaitMin = 1 * time.Millisecond
	conf.RetryWaitMax = 2 * time.Millisecond
	client := NewClient(conf)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&accessCount), "4xx must NOT be retried")
}

func TestClient_Do_ReplaysBodyAcrossRetries(t *testing.T) {
	const payload = "hello-retry-body"
	var attempts int32
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conf := DefaultConfig()
	conf.MaxRetries = 5
	conf.RetryWaitMin = 1 * time.Millisecond
	conf.RetryWaitMax = 2 * time.Millisecond
	client := NewClient(conf)

	req, _ := http.NewRequest("POST", server.URL, bytes.NewBufferString(payload))
	resp, err := client.Do(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	require.Len(t, bodies, 3)
	for i, got := range bodies {
		assert.Equalf(t, payload, got, "attempt %d received a truncated/empty body", i+1)
	}
}

func TestClient_Do_ReturnsSentinelOnRetryExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	conf := DefaultConfig()
	conf.MaxRetries = 1
	conf.RetryWaitMin = 1 * time.Millisecond
	conf.RetryWaitMax = 2 * time.Millisecond
	client := NewClient(conf)

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMaxRetriesExceeded)
}

func TestClient_calculateBackoff_CapsAtMax(t *testing.T) {
	conf := ClientConfig{
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 5 * time.Second,
	}
	c := NewClient(conf)

	for attempt := 5; attempt <= 30; attempt++ {
		got := c.calculateBackoff(attempt)
		upper := time.Duration(float64(conf.RetryWaitMax) * 1.1)
		assert.LessOrEqualf(t, got, upper, "backoff at attempt %d exceeded capped jitter ceiling", attempt)
		assert.Positive(t, got)
	}
}
