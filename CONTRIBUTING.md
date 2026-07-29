# コントリビューションガイド

gloudia への貢献を歓迎します。

## はじめに

- **脆弱性の報告は Issue ではなく [SECURITY.md](SECURITY.md) の手順**でお願いします。
- 大きな変更を行う前に、まず Issue で方針を相談してください。実装後に設計方針が合わずお互いの労力が無駄になるのを防げます。

## 開発環境

- Go 1.25.0 以上

```bash
git clone https://github.com/golaboratory/gloudia.git
cd gloudia
go mod download
go test ./...
```

## 変更を送る前に

以下がすべて通ることを確認してください。CI でも同じチェックを実行します。

```bash
go mod tidy -diff          # go.mod / go.sum が整理済みか
go vet ./...               # 静的解析
go test -race ./...        # レースディテクタ付きテスト
gofmt -l .                 # 出力が空であること
golangci-lint run ./...    # リンター (0 issues であること)
```

`golangci-lint` は CI と同じバージョンを使ってください。設定は `.golangci.yml` にあります。

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
```

> [!NOTE]
> `.golangci.yml` には除外ルールが理由付きで列挙されています。
> 新たに除外を追加する場合は、**なぜ対応不要なのか**をコメントで明記してください。
> 現在の既知の除外項目は、テストコードの `errcheck` と
> `storage/sakura.go` の aws-sdk-go-v2 非推奨 API（`feature/s3/transfermanager` への移行が別途必要）です。

## コーディング規約

### 命名

| 対象 | 規約 | 例 |
|---|---|---|
| エクスポート | PascalCase | `TokenMaker`, `NewClient` |
| 非エクスポート | camelCase | `checkOrigin` |
| コンストラクタ | `New*` プレフィックス | `NewTokenMaker` |
| 設定構造体 | `Config` または `*Config` + `DefaultConfig()` ファクトリ | `ClientConfig` / `DefaultConfig()` |
| センチネルエラー | `Err*` プレフィックス | `ErrInvalidKeySize` |
| レシーバ | 1〜2 文字 | `c`, `h`, `w` |

### エラーハンドリング

`fmt.Errorf` / `errors.New` ではなく [`ergo`](https://github.com/newmo-oss/ergo) を使用します。

```go
ergo.NewSentinel("expected condition")   // センチネルエラーの定義
ergo.Wrap(err, "context message")        // 既存エラーのラップ
```

動的な文脈情報を付けつつ `errors.Is()` を機能させたい場合は `fmt.Errorf("%w: %v", ErrXxx, detail)` を使います。

### ロギング

標準ライブラリの `log/slog` を使用します。context がある場合は context 付きの変種を使ってください。

```go
slog.InfoContext(ctx, "message", slog.String("key", "value"))
```

**機密情報をログに出さないでください。** パスワード・トークン・鍵・資格情報を含む URL は
マスクしてから出力します（`middleware.redactSensitive` / `httpclient.redactURL` を参照）。

### テスト

- フレームワーク: [`testify`](https://github.com/stretchr/testify)（`assert` + `require`）
- スタイル: `t.Run` によるテーブルドリブンテスト
- 前提条件の検証には `require`、結果の検証には `assert`
- ファイル名は対象ソースに対応させる（`foo.go` → `foo_test.go`）
- 共有フィクスチャは `_testdata/` に置く

```go
func TestExample(t *testing.T) {
    t.Run("case name", func(t *testing.T) {
        result, err := SomeFunction(input)
        require.NoError(t, err)
        assert.Equal(t, expected, result)
    })
}
```

**テストデータに実在の個人情報・組織情報・資格情報を含めないでください。**
Excel などバイナリのフィクスチャを追加する際は、作成者名や絶対パスなどのメタデータを除去してください。

## セキュリティに関わる変更

以下に該当する変更は、PR の説明で必ず明示してください。

- 既定値を緩める変更（fail-closed → fail-open）
- 認証・認可・テナント分離に関わる変更
- ログ出力の追加（機密情報が含まれないことの確認）

本ライブラリは**安全な既定値**を方針としています。利便性のために既定を緩める提案は、
オプトイン（明示的に有効化する API）として実装してください。

## バージョニング

`v0` 系のため、マイナーバージョンで破壊的変更が入ることがあります。
破壊的変更を含む PR では、移行手順を PR 説明に記載してください。

## ライセンス

貢献されたコードは [MIT License](LICENSE) の下で公開されることに同意したものとみなします。
