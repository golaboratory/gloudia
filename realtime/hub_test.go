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

}

// TestHub_MultipleClients_Broadcast は複数クライアントへの同時ブロードキャストをテストします。
func TestHub_MultipleClients_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		hub.register <- client
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	// 複数クライアントを接続
	const numClients = 5
	conns := make([]*websocket.Conn, numClients)
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	for i := 0; i < numClients; i++ {
		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("dial %d failed: %v", i, err)
		}
		defer ws.Close()
		conns[i] = ws
	}

	time.Sleep(100 * time.Millisecond)

	// ブロードキャスト
	msg := []byte("broadcast to all")
	hub.BroadcastToAll(msg)

	// 全クライアントで受信確認
	for i, ws := range conns {
		ws.SetReadDeadline(time.Now().Add(time.Second))
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("client %d read failed: %v", i, err)
		}
		if string(p) != string(msg) {
			t.Errorf("client %d: expected %s, got %s", i, msg, p)
		}
	}
}

// TestHub_Unregister はクライアント切断後にブロードキャストが残存クライアントのみに届くことを確認します。
func TestHub_Unregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		hub.register <- client
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// 2つのクライアントを接続
	ws1, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial ws1 failed: %v", err)
	}
	defer ws1.Close()

	ws2, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial ws2 failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// ws2 を切断 → readPump が unregister を発火
	ws2.Close()
	time.Sleep(100 * time.Millisecond)

	// ブロードキャスト
	hub.BroadcastToAll([]byte("after disconnect"))

	// ws1 は受信できる
	ws1.SetReadDeadline(time.Now().Add(time.Second))
	_, p, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("ws1 read failed: %v", err)
	}
	if string(p) != "after disconnect" {
		t.Errorf("expected 'after disconnect', got %s", p)
	}
}

// TestHub_ConcurrentRegisterUnregister は並行クライアント登録・解除でレースコンディションが発生しないことを確認します。
// go test -race で実行してください。
func TestHub_ConcurrentRegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		hub.register <- client
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// 並行で複数クライアントを接続・切断
	const numClients = 20
	done := make(chan struct{}, numClients)

	for i := 0; i < numClients; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			ws, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				return
			}
			// 短時間接続して切断
			time.Sleep(50 * time.Millisecond)
			ws.Close()
		}()
	}

	// 並行でブロードキャストも実行
	go func() {
		for i := 0; i < 10; i++ {
			hub.BroadcastToAll([]byte("concurrent message"))
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// 全クライアントの完了を待つ
	for i := 0; i < numClients; i++ {
		<-done
	}

	// レースコンディションがなくテストが完了すること自体が成功基準
	time.Sleep(200 * time.Millisecond)
}

// newTenantTestClient はテナント別配信テスト用のクライアント（WebSocket接続なし）を生成します。
func newTenantTestClient(hub *Hub, userID int64, tenantID string, buffer int) *Client {
	return &Client{
		hub:      hub,
		send:     make(chan []byte, buffer),
		userID:   userID,
		tenantID: tenantID,
	}
}

// TestHub_BroadcastToTenant はテナント別配信が対象テナントのクライアントのみに届くことを検証します。
func TestHub_BroadcastToTenant(t *testing.T) {
	tests := []struct {
		name         string
		targetTenant string
		wantTenantA  int // テナントAのクライアントが受信するメッセージ数
		wantTenantB  int // テナントBのクライアントが受信するメッセージ数
	}{
		{name: "テナントAのみに配信される", targetTenant: "tenant-a", wantTenantA: 1, wantTenantB: 0},
		{name: "テナントBのみに配信される", targetTenant: "tenant-b", wantTenantA: 0, wantTenantB: 1},
		{name: "存在しないテナントは誰にも届かない", targetTenant: "tenant-x", wantTenantA: 0, wantTenantB: 0},
		{name: "空のテナントIDは誰にも届かない", targetTenant: "", wantTenantA: 0, wantTenantB: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHub()
			clientA := newTenantTestClient(hub, 1, "tenant-a", 4)
			clientB := newTenantTestClient(hub, 2, "tenant-b", 4)
			hub.clients[clientA] = true
			hub.clients[clientB] = true

			hub.BroadcastToTenant(tt.targetTenant, []byte("hello"))

			if got := len(clientA.send); got != tt.wantTenantA {
				t.Errorf("テナントAの受信数: expected %d, got %d", tt.wantTenantA, got)
			}
			if got := len(clientB.send); got != tt.wantTenantB {
				t.Errorf("テナントBの受信数: expected %d, got %d", tt.wantTenantB, got)
			}
		})
	}
}

// TestHub_BroadcastToTenant_FullBuffer は送信バッファ満杯のクライアントがいても
// ブロックせずに処理が完了することを検証します。
func TestHub_BroadcastToTenant_FullBuffer(t *testing.T) {
	hub := NewHub()
	full := newTenantTestClient(hub, 1, "tenant-a", 0) // バッファ0=常に満杯
	normal := newTenantTestClient(hub, 2, "tenant-a", 4)
	hub.clients[full] = true
	hub.clients[normal] = true

	done := make(chan struct{})
	go func() {
		hub.BroadcastToTenant("tenant-a", []byte("hello"))
		close(done)
	}()

	select {
	case <-done:
		// ブロックせず完了すること
	case <-time.After(time.Second):
		t.Fatal("BroadcastToTenant がブロックしました")
	}
	if got := len(normal.send); got != 1 {
		t.Errorf("正常クライアントの受信数: expected 1, got %d", got)
	}
}

// TestHub_BroadcastToTenant_ConcurrentWithRun はメインループ稼働中の並行テナント配信で
// データ競合（-race で検出）が発生しないことを検証します。
func TestHub_BroadcastToTenant_ConcurrentWithRun(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := newTenantTestClient(hub, 1, "tenant-a", 4)
	hub.register <- client

	// 登録がメインループへ反映されるのを待つ
	deadline := time.Now().Add(time.Second)
	for {
		hub.mu.RLock()
		n := len(hub.clients)
		hub.mu.RUnlock()
		if n == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("クライアント登録が反映されませんでした")
		}
		time.Sleep(10 * time.Millisecond)
	}

	for i := 0; i < 50; i++ {
		hub.BroadcastToTenant("tenant-a", []byte("ping"))
		// 受信分を読み捨ててバッファ満杯を避ける
		select {
		case <-client.send:
		default:
		}
	}

	hub.unregister <- client
	time.Sleep(100 * time.Millisecond)
}

// TestHub_BroadcastToUsers は宛先ユーザー集合に含まれるクライアントのみに配信されることを検証します。
func TestHub_BroadcastToUsers(t *testing.T) {
	tests := []struct {
		name         string
		targetTenant string
		userIDs      []int64
		wantUser1    int // tenant-a / userID=1 の受信数
		wantUser2    int // tenant-a / userID=2 の受信数
		wantOther    int // tenant-b / userID=1 の受信数（テナント越え検証）
	}{
		{name: "宛先ユーザーのみに配信される", targetTenant: "tenant-a", userIDs: []int64{1}, wantUser1: 1, wantUser2: 0, wantOther: 0},
		{name: "複数宛先に配信される", targetTenant: "tenant-a", userIDs: []int64{1, 2}, wantUser1: 1, wantUser2: 1, wantOther: 0},
		{name: "同一ユーザーIDでも別テナントへは届かない", targetTenant: "tenant-b", userIDs: []int64{1}, wantUser1: 0, wantUser2: 0, wantOther: 1},
		{name: "宛先が空なら誰にも届かない", targetTenant: "tenant-a", userIDs: nil, wantUser1: 0, wantUser2: 0, wantOther: 0},
		{name: "空のテナントIDは誰にも届かない", targetTenant: "", userIDs: []int64{1, 2}, wantUser1: 0, wantUser2: 0, wantOther: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHub()
			user1 := newTenantTestClient(hub, 1, "tenant-a", 4)
			user2 := newTenantTestClient(hub, 2, "tenant-a", 4)
			other := newTenantTestClient(hub, 1, "tenant-b", 4)
			hub.clients[user1] = true
			hub.clients[user2] = true
			hub.clients[other] = true

			hub.BroadcastToUsers(tt.targetTenant, tt.userIDs, []byte("hello"))

			if got := len(user1.send); got != tt.wantUser1 {
				t.Errorf("tenant-a/user1 の受信数: expected %d, got %d", tt.wantUser1, got)
			}
			if got := len(user2.send); got != tt.wantUser2 {
				t.Errorf("tenant-a/user2 の受信数: expected %d, got %d", tt.wantUser2, got)
			}
			if got := len(other.send); got != tt.wantOther {
				t.Errorf("tenant-b/user1 の受信数: expected %d, got %d", tt.wantOther, got)
			}
		})
	}
}

// TestHub_BroadcastToUsers_MultipleConnections は同一ユーザーの複数接続（複数タブ・複数端末）
// すべてに配信されることを検証します。
func TestHub_BroadcastToUsers_MultipleConnections(t *testing.T) {
	hub := NewHub()
	tab1 := newTenantTestClient(hub, 1, "tenant-a", 4)
	tab2 := newTenantTestClient(hub, 1, "tenant-a", 4)
	hub.clients[tab1] = true
	hub.clients[tab2] = true

	hub.BroadcastToUsers("tenant-a", []int64{1}, []byte("hello"))

	if got := len(tab1.send); got != 1 {
		t.Errorf("接続1の受信数: expected 1, got %d", got)
	}
	if got := len(tab2.send); got != 1 {
		t.Errorf("接続2の受信数: expected 1, got %d", got)
	}
}

// TestHub_BroadcastToUsers_FullBuffer は送信バッファ満杯のクライアントがいても
// ブロックせずに処理が完了することを検証します。
func TestHub_BroadcastToUsers_FullBuffer(t *testing.T) {
	hub := NewHub()
	full := newTenantTestClient(hub, 1, "tenant-a", 0) // バッファ0=常に満杯
	normal := newTenantTestClient(hub, 2, "tenant-a", 4)
	hub.clients[full] = true
	hub.clients[normal] = true

	done := make(chan struct{})
	go func() {
		hub.BroadcastToUsers("tenant-a", []int64{1, 2}, []byte("hello"))
		close(done)
	}()

	select {
	case <-done:
		// ブロックせず完了すること
	case <-time.After(time.Second):
		t.Fatal("BroadcastToUsers がブロックしました")
	}
	if got := len(normal.send); got != 1 {
		t.Errorf("正常クライアントの受信数: expected 1, got %d", got)
	}
}

// TestHub_ConnectedUserIDs は接続中ユーザーIDの取得（テナント絞り込み・重複排除）を検証します。
func TestHub_ConnectedUserIDs(t *testing.T) {
	hub := NewHub()
	// tenant-a: userID=1 が 2 接続（複数タブ）、userID=2 が 1 接続 / tenant-b: userID=3
	hub.clients[newTenantTestClient(hub, 1, "tenant-a", 1)] = true
	hub.clients[newTenantTestClient(hub, 1, "tenant-a", 1)] = true
	hub.clients[newTenantTestClient(hub, 2, "tenant-a", 1)] = true
	hub.clients[newTenantTestClient(hub, 3, "tenant-b", 1)] = true

	t.Run("テナント内の接続ユーザーが重複なしで返る", func(t *testing.T) {
		got := hub.ConnectedUserIDs("tenant-a")
		if len(got) != 2 {
			t.Fatalf("ユーザー数: expected 2, got %d (%v)", len(got), got)
		}
		seen := map[int64]bool{}
		for _, id := range got {
			seen[id] = true
		}
		if !seen[1] || !seen[2] {
			t.Errorf("expected {1,2}, got %v", got)
		}
	})

	t.Run("他テナントのユーザーは含まれない", func(t *testing.T) {
		got := hub.ConnectedUserIDs("tenant-b")
		if len(got) != 1 || got[0] != 3 {
			t.Errorf("expected [3], got %v", got)
		}
	})

	t.Run("未接続テナントは空", func(t *testing.T) {
		if got := hub.ConnectedUserIDs("tenant-x"); len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("空のテナントIDは nil", func(t *testing.T) {
		if got := hub.ConnectedUserIDs(""); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}
