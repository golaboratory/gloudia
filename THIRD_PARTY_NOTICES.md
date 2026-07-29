# サードパーティ ライセンス表記

gloudia は以下のオープンソースソフトウェアを利用しています。
各ライセンスの全文は、それぞれのモジュールのリポジトリまたは
ローカルのモジュールキャッシュ (`$(go env GOMODCACHE)/<module>@<version>/LICENSE`) を参照してください。

本ファイルは gloudia のビルドに実際に取り込まれるモジュール
（`go list -deps ./...` に現れるもの）を対象としています。
テスト専用の依存や、依存モジュールがビルドに使用しない推移的依存は含みません。

> **確認結果**: GPL / LGPL / AGPL / MPL / SSPL / BUSL / Commons Clause など、
> MIT ライセンスでの再頒布に制約を課すライセンスの依存は **ありません**。
> 全依存が permissive ライセンス (Apache-2.0 / MIT / BSD / ISC) です。

## Apache License 2.0 に基づく告知

以下のモジュールは Apache License 2.0 で提供され、NOTICE ファイルを同梱しています。
Apache-2.0 第 4 条 (d) に基づき、その存在を明記します。

- `github.com/aws/aws-sdk-go-v2` — NOTICE ファイルはモジュール配布物に同梱されています
- `github.com/aws/smithy-go` — NOTICE ファイルはモジュール配布物に同梱されています
- `github.com/pquerna/otp` — NOTICE ファイルはモジュール配布物に同梱されています
- `google.golang.org/grpc` — NOTICE ファイルはモジュール配布物に同梱されています

---

## ライセンス別一覧（全 68 モジュール）

### Apache-2.0 — 37 モジュール

| モジュール | バージョン | 区分 |
|---|---|---|
| `github.com/aws/aws-sdk-go-v2` | v1.41.5 | direct |
| `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream` | v1.7.8 | indirect |
| `github.com/aws/aws-sdk-go-v2/config` | v1.32.13 | direct |
| `github.com/aws/aws-sdk-go-v2/credentials` | v1.19.13 | direct |
| `github.com/aws/aws-sdk-go-v2/feature/ec2/imds` | v1.18.21 | indirect |
| `github.com/aws/aws-sdk-go-v2/feature/s3/manager` | v1.22.10 | direct |
| `github.com/aws/aws-sdk-go-v2/internal/configsources` | v1.4.21 | indirect |
| `github.com/aws/aws-sdk-go-v2/internal/endpoints/v2` | v2.7.21 | indirect |
| `github.com/aws/aws-sdk-go-v2/internal/ini` | v1.8.6 | indirect |
| `github.com/aws/aws-sdk-go-v2/internal/v4a` | v1.4.22 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding` | v1.13.7 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/internal/checksum` | v1.9.13 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/internal/presigned-url` | v1.13.21 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/internal/s3shared` | v1.19.21 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/s3` | v1.97.3 | direct |
| `github.com/aws/aws-sdk-go-v2/service/signin` | v1.0.9 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/sso` | v1.30.14 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/ssooidc` | v1.35.18 | indirect |
| `github.com/aws/aws-sdk-go-v2/service/sts` | v1.41.10 | indirect |
| `github.com/aws/smithy-go` | v1.24.2 | indirect |
| `github.com/go-logr/logr` | v1.4.3 | indirect |
| `github.com/go-logr/stdr` | v1.2.2 | indirect |
| `github.com/pquerna/otp` | v1.5.0 | direct |
| `github.com/richardlehane/mscfb` | v1.0.6 | indirect |
| `github.com/richardlehane/msoleps` | v1.0.6 | indirect |
| `github.com/sashabaranov/go-openai` | v1.41.2 | direct |
| `go.opentelemetry.io/auto/sdk` | v1.2.1 | indirect |
| `go.opentelemetry.io/otel` | v1.42.0 | direct |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace` | v1.42.0 | indirect |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` | v1.42.0 | direct |
| `go.opentelemetry.io/otel/metric` | v1.42.0 | indirect |
| `go.opentelemetry.io/otel/sdk` | v1.42.0 | direct |
| `go.opentelemetry.io/otel/trace` | v1.42.0 | direct |
| `go.opentelemetry.io/proto/otlp` | v1.10.0 | indirect |
| `google.golang.org/genproto/googleapis/api` | v0.0.0-20260319201613-d00831a3d3e7 | indirect |
| `google.golang.org/genproto/googleapis/rpc` | v0.0.0-20260319201613-d00831a3d3e7 | indirect |
| `google.golang.org/grpc` | v1.79.3 | indirect |

### MIT — 17 モジュール

| モジュール | バージョン | 区分 |
|---|---|---|
| `aidanwoods.dev/go-paseto` | v1.6.0 | direct |
| `aidanwoods.dev/go-result` | v0.3.1 | indirect |
| `github.com/boombuler/barcode` | v1.1.0 | indirect |
| `github.com/cenkalti/backoff/v5` | v5.0.3 | indirect |
| `github.com/cespare/xxhash/v2` | v2.3.0 | indirect |
| `github.com/danielgtaylor/huma/v2` | v2.37.3 | direct |
| `github.com/dgryski/go-rendezvous` | v0.0.0-20200823014737-9f7001d12a5f | indirect |
| `github.com/go-chi/cors` | v1.2.2 | direct |
| `github.com/jackc/pgpassfile` | v1.0.0 | indirect |
| `github.com/jackc/pgservicefile` | v0.0.0-20240606120523-5a60cdf6a761 | indirect |
| `github.com/jackc/pgx/v5` | v5.9.1 | direct |
| `github.com/jackc/puddle/v2` | v2.2.2 | indirect |
| `github.com/kelseyhightower/envconfig` | v1.4.0 | direct |
| `github.com/newmo-oss/ergo` | v0.1.2 | direct |
| `github.com/newmo-oss/go-caller` | v0.1.0 | indirect |
| `github.com/tiendc/go-deepcopy` | v1.7.2 | indirect |
| `go.uber.org/atomic` | v1.11.0 | indirect |

### BSD-3-Clause — 11 モジュール

| モジュール | バージョン | 区分 |
|---|---|---|
| `github.com/google/uuid` | v1.6.0 | indirect |
| `github.com/grpc-ecosystem/grpc-gateway/v2` | v2.28.0 | indirect |
| `github.com/xuri/efp` | v0.0.1 | indirect |
| `github.com/xuri/excelize/v2` | v2.10.1 | direct |
| `github.com/xuri/nfp` | v0.0.2-0.20250530014748-2ddeb826f9a9 | indirect |
| `golang.org/x/crypto` | v0.49.0 | direct |
| `golang.org/x/net` | v0.52.0 | indirect |
| `golang.org/x/sync` | v0.20.0 | indirect |
| `golang.org/x/sys` | v0.42.0 | indirect |
| `golang.org/x/text` | v0.35.0 | indirect |
| `google.golang.org/protobuf` | v1.36.11 | indirect |

### BSD-2-Clause — 3 モジュール

| モジュール | バージョン | 区分 |
|---|---|---|
| `github.com/go-redis/redis_rate/v10` | v10.0.1 | direct |
| `github.com/gorilla/websocket` | v1.5.3 | direct |
| `github.com/redis/go-redis/v9` | v9.18.0 | direct |

---

## 参考にしたコード

`datetime/calendar/jp` の太陰太陽暦（旧暦）変換ロジックおよび内部テーブルは、
[.NET (dotnet/runtime)](https://github.com/dotnet/runtime) の
`System.Globalization.JapaneseLunisolarCalendar` / `EastAsianLunisolarCalendar`
の実装を参考に Go へ移植したものです。

```
Copyright (c) .NET Foundation and Contributors
Licensed under the MIT License.
```

---

## 更新方法

依存関係を変更した場合は、以下で一覧を再確認してください。

```bash
go list -deps ./... | xargs go list -f '{{if .Module}}{{.Module.Path}} {{.Module.Version}}{{end}}' | sort -u
```
