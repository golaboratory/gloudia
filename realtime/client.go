package realtime

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// ピアへの書き込み待ち時間
	writeWait = 10 * time.Second

	// ピアからのPong待ち時間
	pongWait = 60 * time.Second

	// ピアへのPing送信間隔 (pongWaitより短くする必要がある)
	pingPeriod = (pongWait * 9) / 10

	// 最大メッセージサイズ
	maxMessageSize = 512
)

// Client は接続中のユーザーとHubの仲介役です。
type Client struct {
	hub *Hub

	// WebSocket接続
	conn *websocket.Conn

	// メッセージ送信バッファ
	send chan []byte

	// 認証情報 (誰の接続か)
	userID   int64
	tenantID string
}

// readPump はWebSocketからの読み込みを処理します。
// 主にPing/Pongの維持や、クライアントからのメッセージ受信（今回は受信要件が薄いため省略可）を行います。
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("WebSocket error", "error", err)
			}
			break
		}
		// クライアントからのメッセージをHubへ送る場合はここに記述
		// c.hub.inbound <- message
	}
}

// writePump はHubから送られてきたメッセージをWebSocketへ書き込みます。
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hubがチャネルを閉じた（切断要求）
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 溜まっているメッセージがあれば一度に送る
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Ping送信
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
