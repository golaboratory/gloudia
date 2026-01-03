package realtime

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHub_Run_Broadcast(t *testing.T) {
	// 1. Hubの作成と起動
	hub := NewHub()
	go hub.Run()

	// 2. WebSocketサーバーの立ち上げ (httptest)
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		// Client作成と登録
		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan []byte, 256),
			userID: 123,
		}
		hub.register <- client

		// 書き込みポンプ（メインのテストでは読み込みポンプは必須ではないが、Close処理などのために起動しておくと良い）
		// ここではシンプルに、サーバー側からクライアントへメッセージを送るテストなので、writePumpを動かす
		go client.writePump()

		// 読み込みポンプも動かしておかないと、Ping/PongやCloseが処理されない場合がある
		go client.readPump()
	}))
	defer server.Close()

	// 3. クライアント(Test側)からの接続 (ws://...)
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer ws.Close()

	// 登録処理が完了するのを少し待つ（本来はコールバック等で同期すべきだが簡易的にSleep）
	time.Sleep(100 * time.Millisecond)

	// 4. Hub経由でメッセージをブロードキャスト
	message := []byte("hello, world")
	hub.BroadcastToAll(message)

	// 5. クライアント側で受信確認
	ws.SetReadDeadline(time.Now().Add(time.Second))
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(p) != string(message) {
		t.Errorf("expected %s, got %s", message, p)
	}

	// 6. 複数クライアントのテストや登録解除のテストも追加可能
}
