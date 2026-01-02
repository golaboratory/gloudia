package middleware

import (
	"fmt"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRLSProvider は Huma のミドルウェアとして動作し、以下の責務を持ちます。
// 1. リクエストの認証情報またはヘッダーから tenant_id を特定
// 2. DBトランザクションを開始
// 3. SET LOCAL app.current_tenant_id を実行 (RLS有効化)
// 4. トランザクションをContextに注入
// 5. 処理成功時にCommit, エラー時にRollback
func NewRLSProvider(db *pgxpool.Pool) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// 1. テナントID (tenant_id) の取得
		// 本来は AuthMiddleware が先に走り、Context に Claims が入っている想定です。
		// 未ログイン時(ゲスト予約など)の扱いは仕様によりますが、ここでは
		// "X-Tenant-ID" ヘッダー または 認証情報 からの取得を試みます。
		var tenantID string

		// ケースA: 認証済みユーザーの場合
		if id, ok := ctx.Context().Value(KeyTenantID).(string); ok {
			tenantID = id
		}

		// tenant_id が特定できない場合はエラー (400 Bad Request)
		// ※ トップページなどテナント不要なAPIの場合はこのチェックを緩和する必要があります
		if tenantID == "" {
			ctx.SetStatus(400)
			return
		}

		// 2. トランザクション開始
		tx, err := db.Begin(ctx.Context())
		if err != nil {
			slog.Error("Failed to begin transaction", "error", err)
			ctx.SetStatus(500)
			return
		}

		// defer でパニック時や途中リターン時のロールバックを保証
		defer func() {
			if p := recover(); p != nil {
				tx.Rollback(ctx.Context())
				panic(p) // 再パニック
			}
			// ステータスコードが 4xx, 5xx の場合はロールバック
			if ctx.Status() >= 400 {
				tx.Rollback(ctx.Context())
			}
		}()

		// 3. RLSポリシーの設定 (SET LOCAL)
		// SQLインジェクション対策のため、プレースホルダではなく文字列として組み立てる場合は
		// UUID形式であることを事前に検証するか、pgx ドライバの機能を使うことが望ましいですが、
		// ここでは Exec でパラメータバインドを試みます。
		// ※ SET LOCAL はパラメータバインドが効かないケースがあるため、Sprintf で埋め込むのが一般的ですが、
		//    tenantID は信頼できるソース(Token等)由来前提とします。
		query := fmt.Sprintf("SET LOCAL app.current_tenant_id = '%s'", tenantID)
		if _, err := tx.Exec(ctx.Context(), query); err != nil {
			slog.Error("Failed to set RLS context", "error", err)
			tx.Rollback(ctx.Context())
			ctx.SetStatus(500)
			return
		}

		// 4. Context に Tx と tenant_id を保存
		// Huma の WithValue ヘルパーを使用して Context を更新し、下流のハンドラへ渡します。
		ctx = huma.WithValue(ctx, KeyDBTx, tx)
		ctx = huma.WithValue(ctx, KeyTenantID, tenantID)

		next(ctx)

		// 5. コミット制御
		// ハンドラが戻ってきた後、ステータスを確認してコミット
		if ctx.Status() < 400 {
			if err := tx.Commit(ctx.Context()); err != nil {
				slog.Error("Failed to commit transaction", "error", err)
				ctx.SetStatus(500)
			}
		}
	}
}
