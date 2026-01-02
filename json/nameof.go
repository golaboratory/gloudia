package json

import (
	"reflect"
	"strings"

	"github.com/newmo-oss/ergo"
)

// NameOf は、ネストされた構造体も含めて探索し、
// 指定されたフィールド変数のポインタからjsonタグを取得します。
func NameOf(rootStructPtr any, targetFieldPtr any) (string, error) {
	// 1. ルート構造体の検証と準備
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
	targetType := vTargetPtr.Elem().Type() // ターゲットの型を取得

	// 3. 再帰探索を開始
	tag, found := findFieldTagRecursive(vRoot, targetAddr, targetType)
	if !found {
		return "", ergo.New("指定されたフィールドが構造体ツリー内に見つかりませんでした")
	}

	// 4. "name,omitempty" から名前部分だけを抽出
	if before, _, ok := strings.Cut(tag, ","); ok {
		return before, nil
	}
	return tag, nil
}

// findFieldTagRecursive は再帰的にフィールドを探索するヘルパー関数です
func findFieldTagRecursive(vVal reflect.Value, targetAddr uintptr, targetType reflect.Type) (string, bool) {
	// vValの型情報を取得
	vType := vVal.Type()

	// フィールド数分ループ
	for i := 0; i < vVal.NumField(); i++ {
		fieldVal := vVal.Field(i)
		fieldType := vType.Field(i)

		// A. アドレス比較（これが探しているフィールドか？）
		// Unexportedフィールドでパニックにならないよう CanAddr をチェック
		if fieldVal.CanAddr() {
			if fieldVal.Addr().Pointer() == targetAddr {
				// アドレスが一致し、かつ型も一致する場合のみ対象とみなす
				// (構造体の先頭フィールドと構造体自体の区別のため)
				if fieldVal.Type() == targetType {
					return fieldType.Tag.Get("json"), true
				}
			}
		}

		// B. ネスト探索（構造体、または構造体へのポインタか？）

		// 探索対象の値を準備
		// ポインタの場合は中身を取り出す (Dereference)
		searchVal := fieldVal
		if searchVal.Kind() == reflect.Ptr {
			if searchVal.IsNil() {
				// nilポインタの中は探せないのでスキップ
				continue
			}
			searchVal = searchVal.Elem()
		}

		// 構造体であれば再帰呼び出し
		if searchVal.Kind() == reflect.Struct {
			// 再帰的に探索
			tag, found := findFieldTagRecursive(searchVal, targetAddr, targetType)
			if found {
				return tag, true
			}
		}
	}

	return "", false
}
