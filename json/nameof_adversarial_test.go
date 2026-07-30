package json_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/golaboratory/gloudia/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNameOf_SentinelErrorIdentity(t *testing.T) {
	type Sample struct {
		A int `json:"a"`
	}
	p := Sample{A: 1}
	unrelated := 99

	t.Run("root not a pointer -> ErrFirstArgMustBeStructPtr", func(t *testing.T) {
		name, err := json.NameOf(p, &p.A)
		require.Error(t, err)
		assert.ErrorIs(t, err, json.ErrFirstArgMustBeStructPtr)
		assert.Empty(t, name)
	})

	t.Run("root pointer to non-struct -> ErrFirstArgMustBeStructPtr", func(t *testing.T) {
		n := 5
		name, err := json.NameOf(&n, &n)
		require.Error(t, err)
		assert.ErrorIs(t, err, json.ErrFirstArgMustBeStructPtr)
		assert.Empty(t, name)
	})

	t.Run("target not a pointer -> ErrSecondArgMustBeFieldPtr", func(t *testing.T) {
		name, err := json.NameOf(&p, p.A)
		require.Error(t, err)
		assert.ErrorIs(t, err, json.ErrSecondArgMustBeFieldPtr)
		assert.Empty(t, name)
	})

	t.Run("target unrelated to root -> ErrFieldNotFound", func(t *testing.T) {
		name, err := json.NameOf(&p, &unrelated)
		require.Error(t, err)
		assert.ErrorIs(t, err, json.ErrFieldNotFound)
		assert.Empty(t, name)
	})
}

func TestNameOf_EmbeddedAndDeepPointerChain(t *testing.T) {
	type EmbeddedBase struct {
		BaseID string `json:"baseId"`
	}
	type WithEmbed struct {
		EmbeddedBase        // true Go embedding (anonymous)
		Own          string `json:"own"`
	}

	t.Run("true embedded struct field tag is promoted", func(t *testing.T) {
		w := &WithEmbed{Own: "x"}
		w.BaseID = "id-1"
		name, err := json.NameOf(w, &w.BaseID)
		require.NoError(t, err)
		assert.Equal(t, "baseId", name)
	})

	t.Run("multi-level pointer chain resolves leaf via recursive fallback", func(t *testing.T) {
		type L2 struct {
			Leaf string `json:"leaf"`
		}
		type L1 struct {
			Next *L2 `json:"next"`
		}
		type Root struct {
			Down *L1 `json:"down"`
		}
		r := &Root{Down: &L1{Next: &L2{Leaf: "v"}}}

		name, err := json.NameOf(r, &r.Down.Next.Leaf)
		require.NoError(t, err)
		assert.Equal(t, "leaf", name)

		// Pointer field itself (intermediate) resolves to its own tag.
		name2, err2 := json.NameOf(r, &r.Down.Next)
		require.NoError(t, err2)
		assert.Equal(t, "next", name2)
	})
}

// TestNameOf_PointerFieldInsideNestedValueStruct は、ネストした値構造体の中に
// ポインタフィールドがある場合（Huma のリクエスト型 `struct{ Body struct{ X *T } }` パターン）の
// 回帰テスト。修正前はポインタフィールドのローカルインデックスをルートに適用して
// `reflect: Field index out of range` で panic していた。
func TestNameOf_PointerFieldInsideNestedValueStruct(t *testing.T) {
	type NestedCorp struct {
		Kind string `json:"kind"`
	}
	type HumaStyleRequest struct {
		Body struct {
			ID   int         `json:"id"`
			Flag *bool       `json:"flag"` // 非構造体ポインタ（探索対象外であることの確認用）
			Corp *NestedCorp `json:"corp"`
		}
	}

	req := &HumaStyleRequest{}
	flag := true
	req.Body.Flag = &flag
	req.Body.Corp = &NestedCorp{Kind: "corporate"}

	t.Run("ネスト値構造体内のポインタ先フィールドを解決できる", func(t *testing.T) {
		name, err := json.NameOf(req, &req.Body.Corp.Kind)
		require.NoError(t, err)
		assert.Equal(t, "kind", name)
	})

	t.Run("ポインタフィールド自体はオフセット一致で解決できる", func(t *testing.T) {
		name, err := json.NameOf(req, &req.Body.Corp)
		require.NoError(t, err)
		assert.Equal(t, "corp", name)
	})

	t.Run("非構造体ポインタの参照先は panic せず見つからない扱いになる", func(t *testing.T) {
		unrelated := 1
		name, err := json.NameOf(req, &unrelated)
		require.Error(t, err)
		assert.ErrorIs(t, err, json.ErrFieldNotFound)
		assert.Empty(t, name)
	})
}

func TestNameOf_ConcurrentSchemaBuild(t *testing.T) {
	type ConcChild struct {
		Value string `json:"concValue"`
	}
	type ConcRoot struct {
		ID    int       `json:"concId"`
		Child ConcChild `json:"concChild"`
	}

	const goroutines = 64
	var wg sync.WaitGroup
	wg.Add(goroutines)
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			r := &ConcRoot{ID: 1}
			n1, e1 := json.NameOf(r, &r.ID)
			if e1 != nil || n1 != "concId" {
				errCh <- fmt.Errorf("ID: got %q err %v", n1, e1)
				return
			}
			n2, e2 := json.NameOf(r, &r.Child.Value)
			if e2 != nil || n2 != "concValue" {
				errCh <- fmt.Errorf("Child.Value: got %q err %v", n2, e2)
				return
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

// TestNameOf_TypedNilFieldPtr は型付き nil ポインタを第二引数に渡した場合に
// panic せず ErrSecondArgMustBeFieldPtr を返すことを検証する。
func TestNameOf_TypedNilFieldPtr(t *testing.T) {
	type S struct {
		Name string `json:"name"`
	}
	s := S{}
	_, err := json.NameOf(&s, (*string)(nil))
	require.Error(t, err)
	assert.ErrorIs(t, err, json.ErrSecondArgMustBeFieldPtr,
		"errors.Is で判定可能であるべき")
}
