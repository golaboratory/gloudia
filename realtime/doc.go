// Package realtime は WebSocket を使用したリアルタイムメッセージング機能を提供します。
//
// # アーキテクチャ
//
// Hub/Client パターンに基づき、以下の3つのコンポーネントで構成されます:
//
//   - Hub: アクティブなクライアント接続を管理し、メッセージのブロードキャストを制御
//   - Client: 個々の WebSocket 接続を管理し、Hub との仲介を行う
//   - Handler: HTTP → WebSocket のアップグレードと認証を処理
//
// # WebSocket プロトコル仕様
//
// 接続ライフサイクル:
//
//  1. クライアントが /ws に接続。認証トークンは Authorization: Bearer <PASETO>
//     ヘッダーで送信することを推奨。?token=<PASETO> クエリパラメータは非推奨
//     （警告ログが出力され、アクセスログや Referer 経由で URL からトークンが
//     漏洩する恐れがある）
//  2. サーバーが PASETO トークンを検証（無効な場合は 401 を返却）
//  3. Origin 検証: 既定では同一オリジンのみ許可（クロスオリジンのブラウザは
//     fail-closed で拒否。Origin ヘッダを送らない非ブラウザクライアントは許可）。
//     WithAllowedOrigins / WithAnyOrigin / 環境変数 WS_ALLOWED_ORIGINS で変更可能
//  4. HTTP → WebSocket アップグレード（101 Switching Protocols）
//  5. Client が Hub に登録され、readPump / writePump ゴルーチンが起動
//  6. 切断時は readPump が Hub への登録解除を実行
//
// メッセージ形式:
//
//   - サーバー → クライアント: TextMessage（UTF-8 テキスト）
//   - BroadcastToAll() は全接続クライアントへの一斉送信で、テナント境界を越える。
//     テナント向け通知は必ず BroadcastToTenant() を、特定ユーザー向けの配信は
//     BroadcastToUsers() を使用する。接続中ユーザーの確認には ConnectedUserIDs()
//     を使用する
//   - 複数メッセージがバッファに溜まっている場合、改行区切りで一括送信
//
// 接続維持:
//
//   - Ping: サーバーから 54 秒間隔で送信
//   - Pong: クライアントは 60 秒以内に応答（タイムアウトで切断）
//   - 読み取りデッドラインは「60 秒の Pong 待ち」と「トークン有効期限」の早い方。
//     トークンの有効期限が到来すると接続は切断される
//   - 最大メッセージサイズ: 512 バイト
//   - 書き込みタイムアウト: 10 秒
package realtime
