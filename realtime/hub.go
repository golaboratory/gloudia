package realtime

import (
	"log/slog"
	"sync"
)

// Hub はアクティブなクライアントの集合を管理し、メッセージをブロードキャストします。
type Hub struct {
	// 登録されたクライアントのマップ (boolはダミー値)
	clients map[*Client]bool

	// クライアントへのブロードキャスト用チャネル
	broadcast chan []byte

	// クライアント登録用チャネル
	register chan *Client

	// クライアント登録解除用チャネル
	unregister chan *Client

	// 排他制御用 (チャネルによる同期を主としつつ、外部から安全にアクセスするために保持)
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
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			slog.Debug("Client registered", "user_id", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				slog.Debug("Client unregistered", "user_id", client.userID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// 全クライアントへメッセージ送信
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 送信バッファがいっぱい、または切断されている場合
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastToAll は全接続クライアントにメッセージを送信します。
// 外部パッケージ(Service等)から呼び出すためのメソッドです。
func (h *Hub) BroadcastToAll(message []byte) {
	h.broadcast <- message
}

// BroadcastToTenant は指定テナント(tenantID)の接続クライアントにのみメッセージを送信します。
// マルチテナント環境で他テナントへの情報漏えいを防ぐため、テナント向け通知は必ず本メソッドを使用してください。
// 送信バッファが満杯のクライアントへの送信はスキップします（切断処理は Run ループ側の Ping/Pong に委ねる）。
func (h *Hub) BroadcastToTenant(tenantID string, message []byte) {
	if tenantID == "" {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.tenantID != tenantID {
			continue
		}
		select {
		case client.send <- message:
		default:
			// バッファ満杯時はスキップ（クローズ/削除は Run ループに委ねる）
		}
	}
}

// BroadcastToUsers は指定テナント(tenantID)のうち、userIDs に含まれるユーザーの接続クライアントにのみ
// メッセージを送信します。RBAC 等でサーバー側が解決した宛先集合に限定して配信するために使用します
// （テナント内の無差別配信による権限外情報の漏えいを防ぐ）。
// 同一ユーザーの複数接続（複数タブ・複数端末）にはすべて配信します。
// 送信バッファが満杯のクライアントへの送信はスキップします（切断処理は Run ループ側の Ping/Pong に委ねる）。
func (h *Hub) BroadcastToUsers(tenantID string, userIDs []int64, message []byte) {
	if tenantID == "" || len(userIDs) == 0 {
		return
	}
	idSet := make(map[int64]struct{}, len(userIDs))
	for _, id := range userIDs {
		idSet[id] = struct{}{}
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.tenantID != tenantID {
			continue
		}
		if _, ok := idSet[client.userID]; !ok {
			continue
		}
		select {
		case client.send <- message:
		default:
			// バッファ満杯時はスキップ（クローズ/削除は Run ループに委ねる）
		}
	}
}

// ConnectedUserIDs は指定テナントで現在接続中のユーザーIDを重複なしで返します（順序は不定）。
// 「接続中＝画面を開いている」とみなし、Push 通知の宛先から除外する判定などに使用します。
func (h *Hub) ConnectedUserIDs(tenantID string) []int64 {
	if tenantID == "" {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	seen := make(map[int64]struct{})
	ids := make([]int64, 0, len(h.clients))
	for client := range h.clients {
		if client.tenantID != tenantID {
			continue
		}
		if _, ok := seen[client.userID]; ok {
			continue
		}
		seen[client.userID] = struct{}{}
		ids = append(ids, client.userID)
	}
	return ids
}
