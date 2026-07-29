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
//  1. クライアントが /ws?token=<PASETO> に接続
//  2. サーバーが PASETO トークンを検証（無効な場合は 401 を返却）
//  3. HTTP → WebSocket アップグレード（101 Switching Protocols）
//  4. Client が Hub に登録され、readPump / writePump ゴルーチンが起動
//  5. 切断時は readPump が Hub への登録解除を実行
//
// メッセージ形式:
//
//   - サーバー → クライアント: TextMessage（UTF-8 テキスト）
//   - BroadcastToAll() で全接続クライアントに一斉送信
//   - 複数メッセージがバッファに溜まっている場合、改行区切りで一括送信
//
// 接続維持:
//
//   - Ping: サーバーから 54 秒間隔で送信
//   - Pong: クライアントは 60 秒以内に応答（タイムアウトで切断）
//   - 最大メッセージサイズ: 512 バイト
//   - 書き込みタイムアウト: 10 秒
package realtime
