package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Notify_UnexpectedStatusNoRetry(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.PostText(context.Background(), "payload")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnexpectedStatus)
	assert.Contains(t, err.Error(), "400")
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "4xx from Slack must not be retried")
}
