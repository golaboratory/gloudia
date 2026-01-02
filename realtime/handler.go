package realtime

import (
	"net/http"

	"github.com/golaboratory/gloudia/auth"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 開発環境用に全オリジンを許可 (本番では厳密に設定することを推奨)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs はWebSocket接続リクエストを処理します。
// Chiルーターなどで `/ws` エンドポイントとして登録します。
func ServeWs(hub *Hub, tokenMaker *auth.TokenMaker, w http.ResponseWriter, r *http.Request) {
	// 1. クエリパラメータからトークンを取得
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// 2. トークン検証
	claims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 3. WebSocketへのアップグレード
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade失敗時のレスポンスはライブラリが行うためログのみ
		return
	}

	// 4. クライアントインスタンスの生成とHubへの登録
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   claims.UserID,
		tenantID: claims.TenantID,
	}
	client.hub.register <- client

	// 5. 読み書きポンプの開始 (ゴルーチン)
	go client.writePump()
	go client.readPump()
}
