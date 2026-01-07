# gloudia

**gloudia** は、`golaboratory` エコシステムのサービス開発を高速化・標準化するための共有 Go ライブラリです。
認証、データベース接続、リアルタイム通信、レポーティング（Excel/PDF）、可観測性（Telemetry）など、マイクロサービス開発に必要な共通機能をモジュールとして提供します。

## 特徴

gloudia は以下の主要な機能コンポーネントを含んでいます：

- **Authentication (`auth`)**: Paseto トークン管理、OTP（ワンタイムパスワード）生成・検証。
- **Infrastructure (`infra`)**: `pgx` を使用した PostgreSQL 接続プール、redis クライアントの管理。
- **Realtime (`realtime`)**: WebSocket を利用したリアルタイムメッセージングのための Hub および Client 実装。
- **API & Middleware (`api`, `middleware`)**: Huma フレームワーク統合、CORS、ロギング、認証、レート制限、テナント解決、RLS 制御などの 高度な HTTP ミドルウェア群。
- **Reporting (`reporting`)**:
  - `excel`: `excelize` をベースとした Excel ファイル生成。
  - `pdf`: Gotenberg API を利用した Excel to PDF 変換クライアント。
- **Storage (`storage`)**: ファイルストレージ操作の抽象化レイヤー（現在はローカルファイルシステムに対応）。
- **Network (`net`)**:
  - `httpclient`: リトライ機能（Exponential Backoff）、ロギング、Context 制御を備えた堅牢な HTTP クライアント。
  - `mail`: SMTP メール送信ユーティリティ。
- **Telemtry (`telemetry`)**: OpenTelemetry を利用した分散トレーシングの初期化と HTTP ミドルウェア、OTLP Exporter サポート。
- **Security (`security/crypto`)**: AES-256-GCM を用いたデータの暗号化・復号ヘルパー。
- **Worker (`worker`)**: データベースをキューとして利用するシンプルなバックグラウンドワーカーフレームワーク。
- **Utilities**:
  - `datetime`: 六曜計算を含む日本の日付・時刻処理。
  - `environment`: `envconfig` を用いた型安全な環境変数ロード。
  - `json`: JSON 操作ヘルパー。

## 要件

- Go 1.25.0 以上

## インストール

プロジェクトの `go.mod` に追加するには、以下のコマンドを実行してください：

```bash
go get github.com/golaboratory/gloudia
```

## 使い方

gloudia はモジュール化されており、必要な機能だけをインポートして利用できます。

### 使用例: OTLP トレーシングの初期化

```go
import (
    "context"
    "log"
    "github.com/golaboratory/gloudia/telemetry"
)

func main() {
    ctx := context.Background()
    // Jaeger (OTLP) へ送信するエクスポーターを作成
    exporter, err := telemetry.NewOTLPExporter(ctx, "localhost:4318", true)
    if err != nil {
        log.Fatal(err)
    }

    // TracerProvider を初期化
    shutdown, err := telemetry.InitTracerProvider("my-service", exporter)
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)
}
```

### 使用例: 堅牢な HTTP クライアント

```go
import (
    "github.com/golaboratory/gloudia/net/httpclient"
)

func main() {
    // リトライ機能付きクライアントの作成
    client := httpclient.NewClient(httpclient.DefaultConfig())

    // 通常の http.Client と同様に使用可能
    resp, err := client.Do(req)
}
```

## ディレクトリ構成

- `api/`: API 定義・レスポンス型
- `auth/`: 認証ロジック (Paseto, OTP)
- `datetime/`: 日付・時刻処理
- `environment/`: 環境変数管理
- `infra/`: DB・Redis インフラ
- `json/`: JSON ユーティリティ
- `middleware/`: HTTP ミドルウェア (Auth, Log, RateLimit, Tenant, RLS)
- `net/`: ネットワーク (HTTP Client, Mail)
- `realtime/`: WebSocket・リアルタイム通信
- `reporting/`: 帳票 (Excel, PDF)
- `security/`: セキュリティ (Crypto)
- `storage/`: ストレージ抽象化
- `telemetry/`: 可観測性 (OpenTelemetry)
- `worker/`: バックグラウンドワーカー

## ライセンス

このプロジェクトは [MIT License](LICENSE) の下でライセンスされています。
詳細については `LICENSE` ファイルを参照してください。
