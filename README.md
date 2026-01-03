# gloudia

**gloudia** は、`golaboratory` エコシステムのサービス開発を高速化・標準化するための共有 Go ライブラリです。
認証、データベース接続、リアルタイム通信、レポーティング（Excel）など、マイクロサービス開発に必要な共通機能をモジュールとして提供します。

## 特徴

gloudia は以下の主要な機能コンポーネントを含んでいます：

- **Authentication (`auth`)**: Paseto トークン管理、OTP（ワンタイムパスワード）生成・検証。
- **Infrastructure (`infra`)**: `pgx` を使用した PostgreSQL 接続プール、redis クライアントの管理。
- **Realtime (`realtime`)**: WebSocket を利用したリアルタイムメッセージングのための Hub および Client 実装。
- **API & Middleware (`api`, `middleware`)**: Huma フレームワーク統合、CORS、ロギングなどの HTTP ミドルウェア。
- **Reporting (`reporting`)**: `excelize` をベースとした Excel ファイル生成・操作ユーティリティ。
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

各パッケージはモジュール化されており、必要な機能だけをインポートして利用できます。

### 使用例: 日付ユーティリティ

```go
import (
    "fmt"
    "github.com/golaboratory/gloudia/datetime"
)

func main() {
    // 六曜などの計算や日付操作
    rokuyo := datetime.GregorianDateToRokuyoString(2026, 1, 1)
    fmt.Println(rokuyo)
}
```

### 使用例: WebSocket Hub

```go
import (
    "github.com/golaboratory/gloudia/realtime"
)

func main() {
    // WebSocket Hubの初期化と起動
    hub := realtime.NewHub()
    go hub.Run()

    // ... ハンドラ内でクライアントアップグレード ...
}
```

## ディレクトリ構成

- `api/`: API 定義・統合
- `auth/`: 認証ロジック
- `datetime/`: 日付・時刻処理
- `environment/`: 環境変数管理
- `infra/`: DB・Redis インフラ
- `middleware/`: HTTP ミドルウェア
- `realtime/`: WebSocket・リアルタイム通信
- `reporting/`: 帳票・Excel 出力
- `worker/`: バックグラウンドワーカー

## ライセンス

このプロジェクトは [MIT License](LICENSE) の下でライセンスされています。
詳細については `LICENSE` ファイルを参照してください。
