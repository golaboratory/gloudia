package middleware

type contextKey string

const (
	// KeyClaims は認証トークンから抽出したユーザー情報(Claims)を保持します
	KeyClaims contextKey = "claims"

	// KeyTenantID はテナントID(UUID)を保持します
	KeyTenantID contextKey = "tenant_id"

	// KeyDBTx はミドルウェアで開始したデータベーストランザクション (pgx.Tx) を保持します
	KeyDBTx contextKey = "db_tx"

	// KeyTenantDomeinName はテナント名を保持します
	KeyTenantDomainName contextKey = "tenant_domain_name"

	// KeyTenantHost はテナント解決で検証済みのホスト名（ポート除去済み、X-Forwarded-Host は
	// 信頼プロキシ経由のみ採用）を保持します。リクエスト由来の Origin を信用せずに
	// サーバ側でリンク等のベースURLを組み立てる用途に使用します。
	KeyTenantHost contextKey = "tenant_host"
)
