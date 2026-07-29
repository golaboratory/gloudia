package realtime

import (
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/golaboratory/gloudia/auth"
	"github.com/gorilla/websocket"
)

// extractToken はWebSocket接続リクエストから認証トークンを取得します。
// 取得の優先順位:
//  1. Authorization: Bearer <token> ヘッダー (推奨。ログや Referer に漏洩しない)
//  2. ?token=<token> クエリパラメータ (非推奨。後方互換のため残置)
//
// クエリパラメータでのトークン送信は、リバースプロキシのアクセスログ・ブラウザ履歴・
// Referer ヘッダー等に残留し漏洩する恐れがあるため (CWE-598)、利用時は警告を出力します。
// 呼び出し側は可能な限り Authorization ヘッダーへ移行してください。
func extractToken(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		fields := strings.Fields(authHeader)
		if len(fields) >= 2 && strings.EqualFold(fields[0], "bearer") {
			return fields[1]
		}
	}

	if token := r.URL.Query().Get("token"); token != "" {
		slog.WarnContext(r.Context(),
			"realtime: authentication token received via URL query string is deprecated and may leak through logs/Referer (CWE-598); use the 'Authorization: Bearer' header instead")
		return token
	}

	return ""
}

// ServeOption は ServeWs の挙動を変更するオプションです。
type ServeOption func(*serveConfig)

// serveConfig は ServeWs のオリジン検証設定です。
type serveConfig struct {
	allowedOrigins []string
	anyOrigin      bool
}

// WithAllowedOrigins は WebSocket 接続を許可する Origin を明示的に指定します。
// 指定した場合、環境変数 WS_ALLOWED_ORIGINS より優先されます。
// 比較はスキームとホストを含む完全一致（大文字小文字は区別しない）です。
//
//	realtime.ServeWs(hub, maker, w, r,
//	    realtime.WithAllowedOrigins("https://app.example.com", "https://admin.example.com"))
func WithAllowedOrigins(origins ...string) ServeOption {
	return func(c *serveConfig) {
		for _, o := range origins {
			if o = strings.TrimSpace(o); o != "" {
				c.allowedOrigins = append(c.allowedOrigins, o)
			}
		}
	}
}

// WithAnyOrigin は任意の Origin からの接続を許可します。
//
// これはクロスサイト WebSocket ハイジャック (CSWSH) に対する防御を無効化します。
// 認証を Cookie ではなくトークンで行っており、かつ任意オリジンからの接続を
// 意図的に受け入れる場合にのみ使用してください。
func WithAnyOrigin() ServeOption {
	return func(c *serveConfig) { c.anyOrigin = true }
}

// resolveServeConfig はオプションと環境変数から実効設定を組み立てます。
// オプションで許可オリジンが指定されなかった場合に限り、環境変数
// WS_ALLOWED_ORIGINS（カンマ区切り）を参照します。
func resolveServeConfig(opts []ServeOption) serveConfig {
	var c serveConfig
	for _, opt := range opts {
		opt(&c)
	}
	if len(c.allowedOrigins) > 0 || c.anyOrigin {
		return c
	}
	if env := strings.TrimSpace(os.Getenv("WS_ALLOWED_ORIGINS")); env != "" {
		for _, o := range strings.Split(env, ",") {
			if o = strings.TrimSpace(o); o != "" {
				c.allowedOrigins = append(c.allowedOrigins, o)
			}
		}
	}
	return c
}

// checkOrigin は WebSocket ハンドシェイクの Origin を検証します。
//
// 既定（許可リスト未設定）では、リクエスト先ホストと同一オリジンからの接続のみを
// 許可します（fail-closed）。これは gorilla/websocket の既定と同じ挙動です。
// クロスオリジン接続を許可するには、WithAllowedOrigins または環境変数
// WS_ALLOWED_ORIGINS で許可リストを明示してください。
//
// Origin ヘッダを送出しないクライアント（ブラウザ以外）は許可します。ブラウザは
// WebSocket ハンドシェイクで必ず Origin を送出するため、これによってクロスサイト
// WebSocket ハイジャック（CSWSH）に対する防御は損なわれません。
func checkOrigin(r *http.Request, c serveConfig) bool {
	if c.anyOrigin {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		// Origin ヘッダを持たない非ブラウザクライアントは許可する。
		return true
	}

	if len(c.allowedOrigins) > 0 {
		for _, o := range c.allowedOrigins {
			if strings.EqualFold(o, origin) {
				return true
			}
		}
		slog.WarnContext(r.Context(), "realtime: rejected WebSocket connection from disallowed origin",
			slog.String("origin", origin))
		return false
	}

	// 既定: 同一オリジンのみ許可する。
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		slog.WarnContext(r.Context(), "realtime: rejected WebSocket connection with malformed Origin",
			slog.String("origin", origin))
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	slog.WarnContext(r.Context(),
		"realtime: rejected cross-origin WebSocket connection; set WS_ALLOWED_ORIGINS or pass realtime.WithAllowedOrigins to permit it",
		slog.String("origin", origin), slog.String("host", r.Host))
	return false
}

// newUpgrader は指定された設定でオリジン検証を行う Upgrader を生成します。
func newUpgrader(c serveConfig) *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return checkOrigin(r, c)
		},
	}
}

// ServeWs はWebSocket接続リクエストを処理します。
// Chiルーターなどで `/ws` エンドポイントとして登録します。
//
// オリジン検証は既定で同一オリジンのみを許可します（fail-closed）。
// クロスオリジン接続を受け付ける場合は WithAllowedOrigins、または環境変数
// WS_ALLOWED_ORIGINS で許可リストを明示してください。
func ServeWs(hub *Hub, tokenMaker *auth.TokenMaker, w http.ResponseWriter, r *http.Request, opts ...ServeOption) {
	// 1. トークンの取得 (Authorization ヘッダー優先、クエリは非推奨フォールバック)
	token := extractToken(r)
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// 2. トークン検証（有効期限も取得し、接続のハードデッドラインに使用する）
	claims, tokenExpiry, err := tokenMaker.VerifyTokenWithExpiry(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 3. WebSocketへのアップグレード
	upgrader := newUpgrader(resolveServeConfig(opts))
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade失敗時のレスポンスはライブラリが行うためログのみ
		return
	}

	// 4. クライアントインスタンスの生成とHubへの登録
	client := &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		userID:      claims.UserID,
		tenantID:    claims.TenantID,
		tokenExpiry: tokenExpiry,
	}
	client.hub.register <- client

	// 5. 読み書きポンプの開始 (ゴルーチン)
	go client.writePump()
	go client.readPump()
}
