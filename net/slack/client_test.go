package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Notify(t *testing.T) {
	// Mock Slack Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var payload IncomingWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify payload content
		if payload.Text == "fail" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		err := client.PostText(ctx, "Hello Slack")
		assert.NoError(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		// httpclientの標準リトライ回数を超えて失敗することを確認
		// (httpclient.DefaultConfigを使うためリトライ待ちが発生するがテストでは許容)
		err := client.PostText(ctx, "fail")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max retries reached")
	})
}
