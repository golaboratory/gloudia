package middleware

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uuidPattern は UUID v4 形式を検証するための正規表現です。
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// NewRLSProvider は Huma のミドルウェアとして動作し、以下の責務を持ちます。
// 1. リクエストの認証情報またはヘッダーから tenant_id を特定
// 2. DBトランザクションを開始
// 3. SET LOCAL app.current_tenant_id を実行 (RLS有効化)
// 4. トランザクションをContextに注入
// 5. 処理成功時にCommit, エラー時にRollback
//
// パターン: Huma (api.UseMiddleware で適用)
func NewRLSProvider(db *pgxpool.Pool) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// 1. テナントID (tenant_id) の取得
		// 本来は AuthMiddleware が先に走り、Context に Claims が入っている想定です。
		// 未ログイン時(ゲスト予約など)の扱いは仕様によりますが、ここでは
		// "X-Tenant-ID" ヘッダー または 認証情報 からの取得を試みます。
		var tenantID string

		// ケースA: 認証済みユーザーの場合
		// NewAuthProvider が認証成功時に claims.TenantID を KeyTenantID へ格納するため、
		// 認証済みルートではこの値は検証済み(authoritative)である。
		if id, ok := ctx.Context().Value(KeyTenantID).(string); ok {
			tenantID = id
		}

		// 多層防御: 認証済みクレームが存在する場合は、RLS に使用する tenant_id が
		// クレームの TenantID と一致することを必ず確認する。
		// (AuthProvider → RLSProvider の順序ミスや、ホスト名由来の値が残存した場合でも
		//  クロステナントアクセスを防ぐ。NewAuthProvider は必ず NewRLSProvider より前に
		//  適用すること。)
		if claims, err := GetClaims(ctx.Context()); err == nil {
			if claims.TenantID == "" || claims.TenantID != tenantID {
				slog.Error("RLS tenant_id does not match authenticated claims",
					"claims_tenant_id", claims.TenantID, "rls_tenant_id", tenantID)
				ctx.SetStatus(403)
				return
			}
		}

		// tenant_id が特定できない場合はエラー (400 Bad Request)
		// ※ トップページなどテナント不要なAPIの場合はこのチェックを緩和する必要があります
		if tenantID == "" {
			ctx.SetStatus(400)
			return
		}

		// SQLインジェクション防止: tenant_id が UUID 形式であることを検証する
		if !uuidPattern.MatchString(tenantID) {
			slog.Error("Invalid tenant_id format", "tenant_id", tenantID)
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

		// コミット成功時のみ true。defer はこのフラグを見てロールバック要否を判定する。
		committed := false

		// defer でパニック時や途中リターン時のロールバックを保証する。
		// コミットが成功しなかったすべてのケース（4xx/5xx・ステータス未設定(0)・
		// 1xx・途中 return・panic）で確実にロールバックする。
		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback(ctx.Context())
				panic(p) // 再パニック
			}
			if !committed {
				_ = tx.Rollback(ctx.Context())
			}
		}()

		// 3. RLSポリシーの設定 (SET LOCAL)
		// SET LOCAL はパラメータバインドが効かないため Sprintf で埋め込みますが、
		// 上記の UUID 形式バリデーションにより安全性を担保しています。
		query := fmt.Sprintf("SET LOCAL app.current_tenant_id = '%s'", tenantID)
		if _, err := tx.Exec(ctx.Context(), query); err != nil {
			slog.Error("Failed to set RLS context", "error", err)
			ctx.SetStatus(500)
			return // committed=false のため defer がロールバックする
		}

		// 4. Context に Tx と tenant_id を保存
		// Huma の WithValue ヘルパーを使用して Context を更新し、下流のハンドラへ渡します。
		ctx = huma.WithValue(ctx, KeyDBTx, tx)
		ctx = huma.WithValue(ctx, KeyTenantID, tenantID)

		next(ctx)

		// 5. コミット制御
		// 成功ステータス(2xx/3xx)のみコミットする。ステータス未設定(0)・1xx・4xx/5xx は
		// 失敗とみなし、committed=false のまま defer がロールバックする。
		//
		// 既知の制限 (commit-after-response): Huma ハンドラはレスポンス本文を next(ctx)
		// 内で下層の ResponseWriter へ書き込むため、この Commit はレスポンス送出後に
		// 実行される。Commit が失敗しても 2xx 応答は既にクライアントへ送出済みであり、
		// 直後の SetStatus(500) はヘッダ確定後となり反映されない（クライアントは成功
		// 応答を受け取るが DB はロールバックされる、というズレが生じうる）。厳密な
		// 一貫性が必要な場合は、レスポンスをバッファリングし Commit 成功後にフラッシュ
		// する実装を利用側で検討すること。
		s := ctx.Status()
		if s >= 200 && s < 400 {
			if err := tx.Commit(ctx.Context()); err != nil {
				slog.Error("Failed to commit transaction", "error", err)
				ctx.SetStatus(500)
			} else {
				committed = true
			}
		}
	}
}
