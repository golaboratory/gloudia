# gloudia

[![CI](https://github.com/golaboratory/gloudia/actions/workflows/ci.yml/badge.svg)](https://github.com/golaboratory/gloudia/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/golaboratory/gloudia.svg)](https://pkg.go.dev/github.com/golaboratory/gloudia)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**gloudia** は、Go による Web API サービス開発でくり返し必要になる基盤機能をまとめた共有ライブラリです。
認証（PASETO / TOTP）、HTTP ミドルウェア（CORS・レート制限・マルチテナント・RLS）、リアルタイム通信（WebSocket）、
帳票（Excel / PDF）、ストレージ抽象化、分散トレーシングなどを、必要なパッケージだけ選んで利用できます。

> [!WARNING]
> **本ライブラリは `v0` 系です。マイナーバージョンで破壊的変更が入ります。**
> 必ずバージョンを固定して利用してください。
>
> **`v0.1.0` は使用しないでください。** v0.1.0 はまったく別の旧コードベースを指しており、
> 認証を迂回可能な既知の問題を含みます。

## 要件

- Go 1.25.0 以上

## インストール

```bash
go get github.com/golaboratory/gloudia@v0.2.0
```

```go
import "github.com/golaboratory/gloudia/auth"
```

---

## パッケージ一覧

| パッケージ | 概要 | 主な型・関数 |
|---|---|---|
| `ai/openai` | OpenAI API クライアント。チャット補完と、任意の型へデコードする画像解析 | `NewClient`, `AnalyzeImage[T]` |
| `api` | Huma 向けの統一 API レスポンス。ハンドラーから直接 `return` できるジェネリックヘルパー | `SetSuccess`, `SetInvalid`, `SetForbidden`, `UnifiedResponseBody` |
| `auth` | PASETO v4.local トークン、bcrypt パスワード、TOTP / 2FA | `TokenMaker`, `HashPassword`, `Setup2FA`, `Verify2FAWithTimeStep` |
| `datetime` | 複数フォーマット対応の日付パースと JST タイムゾーン | `ParseFlexibleDate`, `JST` |
| `datetime/calendar/jp` | 日本の暦。旧暦（太陰太陽暦）・元号・和暦表記・六曜・二十四節気・干支・和風月名（旧暦変換は 1960〜2049 年） | `JapaneseLunisolarCalendar`, `GregorianDateToWarekiString`, `GregorianDateToRokuyoString` |
| `environment` | `envconfig` によるジェネリックな型安全環境変数ロード | `NewEnvValue[T]` |
| `infra` | 接続プール設定付き Redis クライアント初期化 | `NewRedisClient` |
| `json` | 構造体フィールドから JSON タグ名を取得 | `NameOf` |
| `json/diff` | 2 つの JSON 間の変更点抽出（監査ログ用途） | `ComputeDiff`, `ChangePoint` |
| `middleware` | Huma / chi 互換のミドルウェア群。認証・ロギング・CORS・レート制限・テナント解決・RLS・robots | `NewAuthProvider`, `NewCORS`, `NewRedisRateLimiter`, `NewTenantResolution`, `NewRLSProvider` |
| `net/httpclient` | 指数バックオフ（±10% ジッター）リトライ付き HTTP クライアント | `NewClient`, `HTTPDoer` |
| `net/mail` | 日本語対応の SMTP 送信 | `NewSMTPSenderWithConfig`, `Sender` |
| `net/slack` | Slack Incoming Webhook 通知 | `NewClient`, `PostText` |
| `realtime` | WebSocket の Hub / Client。全体・テナント単位・ユーザー指定のブロードキャスト | `NewHub`, `ServeWs`, `Hub.BroadcastToTenant` |
| `reporting/excel` | `excelize` ベースの既存 Excel ファイル操作（読み取り・シートコピー・保存） | `Excel`, `CellPosition` |
| `reporting/pdf` | Gotenberg API を用いた Excel → PDF 変換 | `NewConverter`, `ConvertOptions` |
| `security/crypto` | AES-256-GCM 暗号化・復号 | `NewCryptor` |
| `storage` | ファイルストレージ抽象化。ローカル FS と さくらのクラウド オブジェクトストレージ（S3 互換） | `Storage`, `NewLocalStorage`, `NewSakuraObjectStorage` |
| `telemetry` | OpenTelemetry 分散トレーシングと HTTP ミドルウェア、OTLP エクスポーター | `InitTracerProvider`, `NewOTLPExporter`, `HTTPMiddleware` |
| `worker` | DB をキューとして使うバックグラウンドジョブ処理 | `NewWorker`, `JobProcessor` |

---

## 使い方

### 認証トークン（PASETO）

```go
package main

import (
	"log"
	"time"

	"github.com/golaboratory/gloudia/auth"
)

func main() {
	// 鍵は 32 バイト (hex 64 文字)。必ず環境変数などから供給し、コードに埋め込まないこと。
	// 開発用の鍵生成は auth.GenerateRandomKey() が使えます。
	maker, err := auth.NewTokenMaker(mustGetenv("PASETO_SYMMETRIC_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	token, err := maker.CreateToken(
		42,                                     // userID
		"018f3a0c-1234-7890-abcd-ef0123456789", // tenantID (UUID)
		1,                                      // roleID
		30*time.Minute,                         // 有効期間
	)
	if err != nil {
		log.Fatal(err)
	}

	claims, err := maker.VerifyToken(token)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("user=%d tenant=%s", claims.UserID, claims.TenantID)
}
```

### リトライ付き HTTP クライアント

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/golaboratory/gloudia/net/httpclient"
)

func main() {
	cfg := httpclient.DefaultConfig() // Timeout 30s / 3 回リトライ / 1〜5 秒バックオフ
	client := httpclient.NewClient(cfg)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/api", nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req) // *http.Client と同じ使い勝手
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
```

### WebSocket（リアルタイム配信）

```go
package main

import (
	"log"
	"net/http"

	"github.com/golaboratory/gloudia/auth"
	"github.com/golaboratory/gloudia/realtime"
)

func main() {
	hub := realtime.NewHub()
	go hub.Run()

	maker, err := auth.NewTokenMaker(mustGetenv("PASETO_SYMMETRIC_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 既定では同一オリジンからの接続のみ許可されます。
		// 別オリジンのフロントエンドから接続する場合は明示的に許可します。
		realtime.ServeWs(hub, maker, w, r,
			realtime.WithAllowedOrigins("https://app.example.com"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### OTLP トレーシングの初期化

```go
package main

import (
	"context"
	"log"

	"github.com/golaboratory/gloudia/telemetry"
)

func main() {
	ctx := context.Background()

	exporter, err := telemetry.NewOTLPExporter(ctx, "localhost:4318", true /* insecure */)
	if err != nil {
		log.Fatal(err)
	}

	shutdown, err := telemetry.InitTracerProvider("my-service", exporter)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("tracer shutdown: %v", err)
		}
	}()
}
```

### 統一 API レスポンス（Huma）

```go
type GetUserOutput struct {
	Body struct {
		api.UnifiedResponseBody
		Name string `json:"name"`
	}
}

func GetUser(ctx context.Context, in *GetUserInput) (*GetUserOutput, error) {
	resp := &GetUserOutput{}

	user, err := repo.Find(ctx, in.ID)
	if errors.Is(err, ErrNotFound) {
		return api.SetNotFound(resp, "ユーザーが見つかりません")
	}
	if err != nil {
		return api.SetError(resp, err)
	}

	resp.Body.Name = user.Name
	return api.SetSuccess(resp, "取得しました")
}
```

---

## 環境変数リファレンス

gloudia が **ライブラリ内部で直接読み取る**環境変数は以下の 6 つです。
いずれも未設定でも動作しますが、**セキュリティに関わるものは環境に応じた設定が必須**です。

### セキュリティに関わるもの（設定必須）

#### `TRUSTED_PROXY_CIDRS`

| 項目 | 内容 |
|---|---|
| **用途** | `X-Forwarded-Host` / `X-Forwarded-For` / `X-Real-IP` を信頼してよい**送信元アドレス範囲**をカンマ区切りの CIDR で指定します。ここに含まれない送信元からのこれらのヘッダーは**すべて無視**されます。 |
| **参照パッケージ** | `middleware`（`NewTenantResolution` = テナント解決、`NewRedisRateLimiter` = レート制限のクライアント識別） |
| **既定値** | 未設定（**いかなる送信元も信頼しない**） |
| **設定例** | `TRUSTED_PROXY_CIDRS="10.0.0.0/8,192.168.0.0/16"`<br>`TRUSTED_PROXY_CIDRS="127.0.0.1/32"`（ローカル開発）<br>`TRUSTED_PROXY_CIDRS="10.0.0.0/8"`（ALB / Nginx が VPC 内にいる構成） |
| **未設定時の影響** | リバースプロキシ配下では**サブドメインによるテナント解決が機能しません**（プロキシの Host で解決されます）。またレート制限がプロキシの IP 単位になり、全ユーザーが 1 つのカウンタを共有します。 |
| **注意** | プライベートアドレスを暗黙に信頼する挙動は v0.2.0 で廃止しました。同一プライベート網に第三者のワークロードが同居する環境（Kubernetes・共有 VPC）で成りすましを許すためです。 |

#### `WS_ALLOWED_ORIGINS`

| 項目 | 内容 |
|---|---|
| **用途** | WebSocket ハンドシェイクで接続を許可する `Origin` をカンマ区切りで指定します。スキームを含む完全一致（大文字小文字は区別しない）で判定します。 |
| **参照パッケージ** | `realtime`（`ServeWs`） |
| **既定値** | 未設定（**同一オリジンからの接続のみ許可** = fail-closed） |
| **設定例** | `WS_ALLOWED_ORIGINS="https://app.example.com"`<br>`WS_ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"`<br>`WS_ALLOWED_ORIGINS="http://localhost:5173"`（開発） |
| **未設定時の影響** | API と別オリジンにあるフロントエンドからの WebSocket 接続が拒否されます。 |
| **注意** | コードから指定する `realtime.WithAllowedOrigins(...)` の方が優先されます。`Origin` ヘッダーを送出しない非ブラウザクライアントは常に許可されます（ブラウザは必ず送出するため CSWSH 防御は損なわれません）。 |

#### `CORS_ALLOWED_ORIGINS`

| 項目 | 内容 |
|---|---|
| **用途** | CORS で許可するオリジンをカンマ区切りで**追加**します。ホストサフィックスのワイルドカード（`https://*.example.com`）も指定できます。 |
| **参照パッケージ** | `middleware`（`NewCORS`） |
| **既定値** | 未設定（`http://localhost:5173` = Vite 既定ポートのみ許可） |
| **設定例** | `CORS_ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"`<br>`CORS_ALLOWED_ORIGINS="https://*.example.com"` |
| **未設定時の影響** | 本番ドメインのフロントエンドから API を呼べません。 |
| **注意** | この変数は既定値を**上書きせず追記**します。`AllowCredentials: true` で動作するため、開発用の `http://localhost:5173` は本番でも許可され続けます。これを避けたい場合は `NewCORS()` を使わず `go-chi/cors` を直接構成してください。 |

### 動作調整用（任意）

#### `CRYPT_COST`

| 項目 | 内容 |
|---|---|
| **用途** | パスワードハッシュ化（bcrypt）のコストパラメータ。値を 1 増やすと計算時間が約 2 倍になります。 |
| **参照パッケージ** | `auth`（`HashPassword`）／ `environment`（`GloudiaEnv.CryptCost`） |
| **既定値** | `10`（`bcrypt.DefaultCost`） |
| **設定例** | `CRYPT_COST=12`（推奨。より強固）<br>`CRYPT_COST=4`（テスト高速化用。**本番では使用しないこと**） |
| **有効範囲** | `4`〜`31`。範囲外の値は無視され、既定の `10` が使われます。 |

#### `IS_DEBUG`

| 項目 | 内容 |
|---|---|
| **用途** | デバッグログを有効にします。有効時はアクセスログに**リクエストボディ（最大 1MB）**が追加されます。ボディとクエリ中のパスワード・トークン等は自動でマスクされます。 |
| **参照パッケージ** | `middleware`（`NewLogger`）／ `environment`（`GloudiaEnv.IsDebug`） |
| **既定値** | `false` |
| **設定例** | `IS_DEBUG=true` |
| **注意** | 本番環境では `false` のままにしてください。マスクは既知のキー名パターンに基づく best-effort であり、すべての機密情報を保証するものではありません。 |

#### `REDIS_POOL_SIZE`

| 項目 | 内容 |
|---|---|
| **用途** | Redis クライアントのコネクションプールサイズ。 |
| **参照パッケージ** | `infra`（`NewRedisClient`） |
| **既定値** | `10` |
| **設定例** | `REDIS_POOL_SIZE=50` |
| **注意** | 正の整数のみ有効。解析できない値や 0 以下は無視され既定値が使われます。タイムアウト値（Dial 10s / Read 30s / Write 30s）は現在ハードコードです。 |

### 利用者アプリ側で定義する設定

以下は gloudia が直接読むものではなく、**利用者が `environment.NewEnvValue[T]` で自由に定義する**ものです。

```go
type AppEnv struct {
	PasetoKey  string `envconfig:"PASETO_SYMMETRIC_KEY" required:"true"`
	RedisAddr  string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	GotenbergURL string `envconfig:"GOTENBERG_URL" default:"http://localhost:3000"`
}

env, err := environment.NewEnvValue[AppEnv]("APP") // APP_PASETO_SYMMETRIC_KEY などを読む
```

> [!IMPORTANT]
> 環境変数名にプレフィックスがないため、利用者アプリの環境変数と衝突する可能性があります。
> `environment.NewEnvValue` はプレフィックス付きで利用できますが、上記 6 変数は
> gloudia が固定名で読み取るため、同名の変数を別用途に使わないでください。

---

## セキュリティ既定値

gloudia は **fail-closed（安全側に倒す）** を既定方針としています。

| 対象 | 既定の挙動 | 緩和方法 |
|---|---|---|
| WebSocket の Origin 検証 | 同一オリジンのみ許可 | `realtime.WithAllowedOrigins(...)` / `WS_ALLOWED_ORIGINS` / `realtime.WithAnyOrigin()` |
| プロキシヘッダーの信頼 | 一切信頼しない | `TRUSTED_PROXY_CIDRS` |
| CORS | `http://localhost:5173` のみ | `CORS_ALLOWED_ORIGINS` |
| レート制限（Redis 障害時） | fail-open（通す）＋ ERROR ログ出力 | — |
| ログのマスク | パスワード・トークン等のキーを部分一致でマスク | — |
| HTTP クライアントの URL ログ | ユーザー情報と機密クエリを除去 | `ClientConfig.RedactURLPath` でパスも除去 |
| `storage.LocalStorage` | シンボリックリンクを解決してパストラバーサルを拒否 | — |
| `middleware.NewRLSProvider` | `tenant_id` を UUID 形式で検証、クレームとの一致も確認 | — |

脆弱性を発見した場合は [SECURITY.md](SECURITY.md) の手順に従ってご報告ください。

---

## バージョニング方針

- [セマンティックバージョニング](https://semver.org/lang/ja/) に従いますが、**`v0` 系のため MINOR バージョンで破壊的変更が入ります**。
- 破壊的変更はリリースノートに移行手順とともに記載します。
- 必ず `go get github.com/golaboratory/gloudia@vX.Y.Z` でバージョンを固定してください。
- API が安定した段階で `v1.0.0` を出す予定です。

---

## 開発

```bash
go mod download
go test ./...           # テスト
go test -race ./...     # レースディテクタ付き
go vet ./...            # 静的解析
```

コントリビューションについては [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。

## ライセンス

[MIT License](LICENSE) の下で公開しています。

依存ライブラリのライセンス表記は [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) を参照してください。
すべての依存が permissive ライセンス（Apache-2.0 / MIT / BSD / ISC）であり、コピーレフトの依存はありません。
