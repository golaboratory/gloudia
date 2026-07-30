package json

import (
	"reflect"
	"strings"
	"sync" // Added for memoization

	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrFirstArgMustBeStructPtr は第一引数が構造体ポインタでない場合のエラー
	ErrFirstArgMustBeStructPtr = ergo.NewSentinel("第一引数は構造体へのポインタである必要があります")
	// ErrSecondArgMustBeFieldPtr は第二引数がフィールドポインタでない場合のエラー
	ErrSecondArgMustBeFieldPtr = ergo.NewSentinel("第二引数はフィールドへのポインタである必要があります")
	// ErrFieldNotFound は指定されたフィールドが構造体ツリー内に見つからない場合のエラー
	ErrFieldNotFound = ergo.NewSentinel("指定されたフィールドが構造体ツリー内に見つかりませんでした")
)

// Global cache for structural schema
var schemaCache sync.Map

// cachedField holds the tag and type of a field at a specific offset
type cachedField struct {
	tag string
	typ reflect.Type
}

// pointerInfo holds metadata for fields that are pointers (and thus need runtime traversal)
type pointerInfo struct {
	// path はスキーマのルート型から見たフィールドインデックスのパス。
	// ネストした値構造体をフラット化して収集するため、単一インデックスではなく
	// パスで保持する（ローカルインデックスをルートの Field(i) に適用すると
	// 範囲外アクセスで panic するため）。パスは値構造体のみを経由する。
	path []int
	typ  reflect.Type
}

// typeSchema contains the flattened field map and list of pointer fields for a specific struct type
type typeSchema struct {
	// fields maps the offset (relative to struct start) to a list of potential fields
	// (collisions are possible for 0-size fields or struct/first-field sharing address)
	fields map[uintptr][]cachedField
	// pointers lists fields that are pointers, requiring recursion to resolve
	pointers []pointerInfo
}

// getSchema retrieves or builds the schema for a given type
func getSchema(t reflect.Type) *typeSchema {
	if v, ok := schemaCache.Load(t); ok {
		return v.(*typeSchema)
	}

	s := &typeSchema{
		fields: make(map[uintptr][]cachedField),
	}
	buildSchema(t, 0, nil, s)

	actual, _ := schemaCache.LoadOrStore(t, s)
	return actual.(*typeSchema)
}

// buildSchema recursively explores the struct type to build the flattened schema
func buildSchema(t reflect.Type, baseOffset uintptr, basePath []int, s *typeSchema) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		currOffset := baseOffset + f.Offset
		tag := f.Tag.Get("json")

		// 1. Add current field to the map (candidate for direct match)
		s.fields[currOffset] = append(s.fields[currOffset], cachedField{
			tag: tag,
			typ: f.Type,
		})

		// 2. If it's a pointer, add to pointers list for runtime traversal
		if f.Type.Kind() == reflect.Pointer {
			// basePath との aliasing を避けるため、必ず新しいスライスを確保する
			path := make([]int, len(basePath)+1)
			copy(path, basePath)
			path[len(basePath)] = i
			s.pointers = append(s.pointers, pointerInfo{
				path: path,
				typ:  f.Type,
			})
			continue
		}

		// 3. If it's a struct (embedded or value), recurse to flatten its fields
		if f.Type.Kind() == reflect.Struct {
			path := make([]int, len(basePath)+1)
			copy(path, basePath)
			path[len(basePath)] = i
			buildSchema(f.Type, currOffset, path, s)
		}
	}
}

// NameOf は、ネストされた構造体も含めて探索し、
// 指定されたフィールド変数のポインタからjsonタグの名前部分を取得します。
// Reflectionの結果をキャッシュ（メモ化）して高速化しています。
//
// 引数が不正な場合は ErrFirstArgMustBeStructPtr / ErrSecondArgMustBeFieldPtr を、
// フィールドが見つからない場合は ErrFieldNotFound を返します（errors.Is で判定可能）。
// フィールドは見つかったが json タグが付いていない場合は ("", nil) を返します。
// `json:"-"` のフィールドは "-" を返し、",omitempty" 等のオプションは除去されます。
func NameOf(rootStructPtr any, targetFieldPtr any) (string, error) {
	// 1. ルート構造体の検証
	vRootPtr := reflect.ValueOf(rootStructPtr)
	if vRootPtr.Kind() != reflect.Pointer || vRootPtr.Elem().Kind() != reflect.Struct {
		return "", ErrFirstArgMustBeStructPtr
	}
	vRoot := vRootPtr.Elem()

	// 2. ターゲットフィールドのポインタ検証
	// nil ポインタは実在するフィールドを指し得ないため拒否する
	// (nil のまま Elem().Type() を呼ぶと panic するため、ここで検証する)
	vTargetPtr := reflect.ValueOf(targetFieldPtr)
	if vTargetPtr.Kind() != reflect.Pointer || vTargetPtr.IsNil() {
		return "", ErrSecondArgMustBeFieldPtr
	}
	targetAddr := vTargetPtr.Pointer()
	targetType := vTargetPtr.Elem().Type()

	// 3. スキーマキャッシュを使用した探索
	tag, found := findFieldTagWithSchema(vRoot, targetAddr, targetType)
	if !found {
		return "", ErrFieldNotFound
	}

	// 4. "name,omitempty" から名前部分だけを抽出
	if before, _, ok := strings.Cut(tag, ","); ok {
		return before, nil
	}
	return tag, nil
}

// findFieldTagWithSchema looks up the target using the cached schema and minimal runtime re-traversal
func findFieldTagWithSchema(vRoot reflect.Value, targetAddr uintptr, targetType reflect.Type) (string, bool) {
	// 安全のため CanAddr チェック（通常、Elem() した時点で Addr 可能）
	if !vRoot.CanAddr() {
		return "", false
	}
	rootAddr := vRoot.Addr().Pointer()
	rootSize := vRoot.Type().Size()
	schema := getSchema(vRoot.Type())

	// 1. Check if the target address falls within this struct's memory block
	if targetAddr >= rootAddr && targetAddr < rootAddr+rootSize {
		// Calculate relative offset
		offset := targetAddr - rootAddr

		// Lookup in fields map
		if candidates, ok := schema.fields[offset]; ok {
			for _, cand := range candidates {
				if cand.typ == targetType {
					return cand.tag, true
				}
			}
		}
	}

	// 2. Fallback: Check pointer fields
	// Use schema to identify which fields are pointers
	for _, ptrInfo := range schema.pointers {
		// path は値構造体のみを経由するため、FieldByIndex が中間 nil ポインタで
		// panic することはない
		ptrVal := vRoot.FieldByIndex(ptrInfo.path)

		// If pointer is nil, skip
		if ptrVal.IsNil() {
			continue
		}

		// Recurse: dereference and search inside
		// ptrVal.Elem() is the value pointed to
		elem := ptrVal.Elem()
		if elem.Kind() != reflect.Struct {
			// 構造体以外（*bool 等）の参照先はフィールドを持たないため探索しない
			// （getSchema が非構造体型で NumField を呼び panic するのを防ぐ）
			continue
		}
		if tag, found := findFieldTagWithSchema(elem, targetAddr, targetType); found {
			return tag, true
		}
	}

	return "", false
}
