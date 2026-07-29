package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/golaboratory/gloudia/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetClaims_Success は認証ミドルウェアによって設定された Claims が
// 正しく取得できることを検証する。
//
// リファクタリング保護:
//   - gloudia 側では TenantUUID 抽出ヘルパーを提供しない方針のため、
//     利用側はこの GetClaims() を起点に独自の TenantUUID 抽出を行う。
//   - GetClaims が「正常系で *auth.Claims をそのまま返す」契約を保護することで、
//     利用側のヘルパー実装が安定する。
func TestGetClaims_Success(t *testing.T) {
	expected := &auth.Claims{
		TenantID: "00000000-0000-0000-0000-000000000001",
		UserID:   42,
	}
	ctx := context.WithValue(context.Background(), KeyClaims, expected)

	got, err := GetClaims(ctx)
	require.NoError(t, err)
	assert.Same(t, expected, got, "GetClaims は context にセットされた Claims ポインタをそのまま返すべき")
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", got.TenantID)
	assert.Equal(t, int64(42), got.UserID)
}

// TestGetClaims_NoClaims は context に Claims が存在しない場合に
// ErrUnauthenticated が返ることを保護する。
//
// リファクタリング保護:
//   - 利用側のハンドラーで `errors.Is(err, middleware.ErrUnauthenticated)` による
//     判定が機能することを保証する（フェーズ 1 の `ergo` 統一の前提条件）。
func TestGetClaims_NoClaims(t *testing.T) {
	ctx := context.Background() // KeyClaims を設定しない

	got, err := GetClaims(ctx)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, ErrUnauthenticated),
		"errors.Is(err, ErrUnauthenticated) で判定可能であるべき")
}

// TestGetClaims_NilClaims は context に明示的に nil の *auth.Claims が
// セットされていた場合の振る舞いを保護する。
func TestGetClaims_NilClaims(t *testing.T) {
	var nilClaims *auth.Claims
	ctx := context.WithValue(context.Background(), KeyClaims, nilClaims)

	got, err := GetClaims(ctx)
	require.Error(t, err)
	assert.Nil(t, got, "nil claims が格納されていた場合は nil を返すべき")
	assert.True(t, errors.Is(err, ErrUnauthenticated))
}

// TestGetClaims_WrongType は context.Value の型が *auth.Claims でない場合に
// ErrUnauthenticated が返ることを保護する。
//
// リファクタリング保護:
//   - 認証ミドルウェアの実装変更で意図せず別の型をセットした場合、
//     利用側で型アサーション失敗が起きる前にエラーで弾かれることを保証する。
func TestGetClaims_WrongType(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"string", "not-claims"},
		{"int", 12345},
		{"map", map[string]string{"tenant_id": "abc"}},
		{"struct value (not pointer)", auth.Claims{TenantID: "x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), KeyClaims, tt.value)
			got, err := GetClaims(ctx)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.True(t, errors.Is(err, ErrUnauthenticated),
				"型不一致時も ErrUnauthenticated が返るべき")
		})
	}
}

// TestGetClaims_KeyTenantID_NotConfused は KeyTenantID と KeyClaims が
// 別のコンテキストキーであることを保護する。
//
// リファクタリング保護:
//   - context_keys.go で KeyTenantID と KeyClaims が typed contextKey として分離されている。
//   - KeyTenantID にセットされた値が誤って Claims として読み取られないことを検証する。
//     （context.WithValue の key は文字列ではなく型付き contextKey で区別される）
func TestGetClaims_KeyTenantID_NotConfused(t *testing.T) {
	// KeyTenantID にだけ値を設定し、KeyClaims は未設定
	ctx := context.WithValue(context.Background(), KeyTenantID, "tenant-uuid-string")

	got, err := GetClaims(ctx)
	require.Error(t, err)
	assert.Nil(t, got, "KeyTenantID 経由の値は GetClaims では取得できないべき")
	assert.True(t, errors.Is(err, ErrUnauthenticated))
}

// TestErrUnauthenticated_IsSentinel は ErrUnauthenticated がセンチネルエラーとして
// `errors.Is` で判定可能であることを保護する。
//
// リファクタリング保護:
//   - ergo.NewSentinel で定義された ErrUnauthenticated は同一インスタンス比較ではなく
//     errors.Is による判定が標準。
//   - `ergo` の API 仕様変更時にこの契約が崩れないことを保証する。
func TestErrUnauthenticated_IsSentinel(t *testing.T) {
	// 直接インスタンス比較
	require.NotNil(t, ErrUnauthenticated)

	// errors.Is での自己一致
	assert.True(t, errors.Is(ErrUnauthenticated, ErrUnauthenticated))

	// GetClaims が返すエラーから errors.Is でセンチネルが判定できる
	_, err := GetClaims(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthenticated),
		"GetClaims 起源のエラーは ErrUnauthenticated で判定可能であるべき")
}

// TestGetClaims_ContractForUtilHelpers は利用側で
// 「ExtractTenantUUID(ctx) のようなヘルパー」を実装する際の振る舞い基盤を保護する。
//
// 利用側の典型的なコード例：
//
//	func ExtractTenantUUID(ctx context.Context) (pgtype.UUID, error) {
//	    claims, err := middleware.GetClaims(ctx)
//	    if err != nil {
//	        return pgtype.UUID{}, err  // ErrUnauthenticated がそのまま伝播
//	    }
//	    var id pgtype.UUID
//	    if err := id.Scan(claims.TenantID); err != nil {
//	        return pgtype.UUID{}, fmt.Errorf("invalid tenant uuid: %w", err)
//	    }
//	    return id, nil
//	}
//
// 本テストは上記利用シナリオで GetClaims が安定動作することを保護する。
func TestGetClaims_ContractForUtilHelpers(t *testing.T) {
	t.Run("正常な UUID 文字列", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), KeyClaims, &auth.Claims{
			TenantID: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			UserID:   1,
		})
		claims, err := GetClaims(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, claims.TenantID, "TenantID は利用側で UUID 変換できる文字列であるべき")
	})

	t.Run("空文字 TenantID", func(t *testing.T) {
		// Claims 自体は存在するが TenantID が空 → GetClaims はエラーにせず、
		// 空文字を持った Claims をそのまま返す（UUID 変換失敗は利用側の責任）
		ctx := context.WithValue(context.Background(), KeyClaims, &auth.Claims{
			TenantID: "",
			UserID:   1,
		})
		claims, err := GetClaims(ctx)
		require.NoError(t, err, "空 TenantID でも GetClaims はエラーにしないべき（UUID 検証は呼出側で実施）")
		assert.Equal(t, "", claims.TenantID)
	})
}
