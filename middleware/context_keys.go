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
)
