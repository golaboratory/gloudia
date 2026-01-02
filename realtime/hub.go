package realtime

import (
	"log/slog"
	"sync"
)

// Hub はアクティブなクライアントの集合を管理し、メッセージをブロードキャストします。
type Hub struct {
	// 登録されたクライアントのマップ (boolはダミー値)
	clients map[*Client]bool

	// クライアントからのメッセージ受信用チャネル（必要に応じて実装）
	// inbound chan []byte

	// クライアントへのブロードキャスト用チャネル
	broadcast chan []byte

	// クライアント登録用チャネル
	register chan *Client

	// クライアント登録解除用チャネル
	unregister chan *Client

	// 排他制御用 (mapの競合回避のため、run loopを使用する場合は不要だが、
	// 外部から安全にアクセスするためにRWMutexを持つパターンもある。
	// ここではGo標準のHubパターンに従い、チャネルによる同期を行います)
	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run はHubのメインループを開始します。ゴルーチンとして起動してください。
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			slog.Debug("Client registered", "user_id", client.userID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				slog.Debug("Client unregistered", "user_id", client.userID)
			}

		case message := <-h.broadcast:
			// 全クライアントへメッセージ送信
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 送信バッファがいっぱい、または切断されている場合
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastToAll は全接続クライアントにメッセージを送信します。
// 外部パッケージ(Service等)から呼び出すためのメソッドです。
func (h *Hub) BroadcastToAll(message []byte) {
	h.broadcast <- message
}

// 必要に応じて「特定のテナント(shrine_id)のみに送信」するメソッドを追加実装します。
// その場合、Client構造体にshrine_idを持たせ、Hub側でフィルタリングします。
