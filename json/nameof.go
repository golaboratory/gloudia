package json

import (
	"reflect"
	"strings"
	"sync" // Added for memoization

	"github.com/newmo-oss/ergo"
)

// Global cache for structural schema
var schemaCache sync.Map

// cachedField holds the tag and type of a field at a specific offset
type cachedField struct {
	tag string
	typ reflect.Type
}

// pointerInfo holds metadata for fields that are pointers (and thus need runtime traversal)
// pointerInfo holds metadata for fields that are pointers (and thus need runtime traversal)
type pointerInfo struct {
	index int // Index of the field in the struct
	typ   reflect.Type
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
	buildSchema(t, 0, s)

	actual, _ := schemaCache.LoadOrStore(t, s)
	return actual.(*typeSchema)
}

// buildSchema recursively explores the struct type to build the flattened schema
func buildSchema(t reflect.Type, baseOffset uintptr, s *typeSchema) {
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
		if f.Type.Kind() == reflect.Ptr {
			s.pointers = append(s.pointers, pointerInfo{
				index: i,
				typ:   f.Type,
			})
			continue
		}

		// 3. If it's a struct (embedded or value), recurse to flatten its fields
		if f.Type.Kind() == reflect.Struct {
			buildSchema(f.Type, currOffset, s)
		}
	}
}

// NameOf は、ネストされた構造体も含めて探索し、
// 指定されたフィールド変数のポインタからjsonタグを取得します。
// Reflectionの結果をキャッシュ（メモ化）して高速化しています。
func NameOf(rootStructPtr any, targetFieldPtr any) (string, error) {
	// 1. ルート構造体の検証
	vRootPtr := reflect.ValueOf(rootStructPtr)
	if vRootPtr.Kind() != reflect.Pointer || vRootPtr.Elem().Kind() != reflect.Struct {
		return "", ergo.New("第一引数は構造体へのポインタである必要があります")
	}
	vRoot := vRootPtr.Elem()

	// 2. ターゲットフィールドのポインタ検証
	vTargetPtr := reflect.ValueOf(targetFieldPtr)
	if vTargetPtr.Kind() != reflect.Pointer {
		return "", ergo.New("第二引数はフィールドへのポインタである必要があります")
	}
	targetAddr := vTargetPtr.Pointer()
	targetType := vTargetPtr.Elem().Type()

	// 3. スキーマキャッシュを使用した探索
	tag, found := findFieldTagWithSchema(vRoot, targetAddr, targetType)
	if !found {
		return "", ergo.New("指定されたフィールドが構造体ツリー内に見つかりませんでした")
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

	// 1. Check if the target address falls within this struct's memory block
	if targetAddr >= rootAddr && targetAddr < rootAddr+rootSize {
		// Calculate relative offset
		offset := targetAddr - rootAddr

		// Get schema
		schema := getSchema(vRoot.Type())

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
	schema := getSchema(vRoot.Type())
	for _, ptrInfo := range schema.pointers {
		ptrVal := vRoot.Field(ptrInfo.index) // Access by index

		// If pointer is nil, skip
		if ptrVal.IsNil() {
			continue
		}

		// Recurse: dereference and search inside
		// ptrVal.Elem() is the value pointed to
		if tag, found := findFieldTagWithSchema(ptrVal.Elem(), targetAddr, targetType); found {
			return tag, true
		}
	}

	return "", false
}
