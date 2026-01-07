package realtime

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golaboratory/gloudia/auth"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestServeWs(t *testing.T) {
	// Setup Hub
	hub := NewHub()
	go hub.Run()

	// Setup TokenMaker
	key := auth.GenerateRandomKey()
	maker, err := auth.NewTokenMaker(key)
	assert.NoError(t, err)

	validToken, err := maker.CreateToken(100, "tenantA", 1, time.Minute)
	assert.NoError(t, err)

	t.Run("success connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeWs(hub, maker, w, r)
		}))
		defer server.Close()

		// Connect using websocket client
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + validToken
		ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

		ws.Close()
	})

	t.Run("missing token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeWs(hub, maker, w, r)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// Websocket handshake failure returns error
		assert.Error(t, err)
		// Check response code if available.
		// Note: Dial returns (conn, resp, err). If handshake fails, err is non-nil.
		// Usually if server returns 401, Dial returns BadHandshake error.
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ServeWs(hub, maker, w, r)
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=invalid_token"
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

		assert.Error(t, err)
		if resp != nil {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
