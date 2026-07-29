package diff

import (
	"encoding/json"
	"reflect"
	"sort"

	"github.com/newmo-oss/ergo"
)

// センチネルエラー定義
var (
	// ErrOldJSONUnmarshalFailed は旧JSONのアンマーシャルに失敗した場合のエラー
	ErrOldJSONUnmarshalFailed = ergo.NewSentinel("failed to unmarshal old json")
	// ErrNewJSONUnmarshalFailed は新JSONのアンマーシャルに失敗した場合のエラー
	ErrNewJSONUnmarshalFailed = ergo.NewSentinel("failed to unmarshal new json")
)

// ChangePoint: 1つの変更点を表す構造体
type ChangePoint struct {
	Field    string `json:"field"`     // 変更されたプロパティ名 (ドット記法)
	OldValue any    `json:"old_value"` // 変更前の値
	NewValue any    `json:"new_value"` // 変更後の値
}

// ComputeDiff: 新旧のJSONバイト列を比較して差分リストを返す
func ComputeDiff(oldJson, newJson []byte) ([]ChangePoint, error) {
	var oldMap, newMap map[string]any

	// 1. JSONをMapに展開
	// create時(oldJsonが空)やdelete時(newJsonが空)のハンドリング
	if len(oldJson) > 0 {
		if err := json.Unmarshal(oldJson, &oldMap); err != nil {
			return nil, ergo.Wrap(ErrOldJSONUnmarshalFailed, err.Error())
		}
	}
	if len(newJson) > 0 {
		if err := json.Unmarshal(newJson, &newMap); err != nil {
			return nil, ergo.Wrap(ErrNewJSONUnmarshalFailed, err.Error())
		}
	}

	// 2. 再帰的に比較
	return compareMaps("", oldMap, newMap), nil
}

// compareMaps: Map同士を再帰比較する内部関数
func compareMaps(path string, oldVal, newVal map[string]any) []ChangePoint {
	changes := []ChangePoint{}
	allKeys := make(map[string]struct{})

	// 両方のキーを収集
	for k := range oldVal {
		allKeys[k] = struct{}{}
	}
	for k := range newVal {
		allKeys[k] = struct{}{}
	}

	// ソートされたキーリストを作成
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	// 順序を固定
	sort.Strings(keys)

	for _, key := range keys {
		// comma-ok でキーの存在を判定する。プレーンな index 参照では
		// 「キーが無い」場合も「キーが明示的な null」の場合も nil となり区別できず、
		// null ⇄ キー削除/追加 の変更を取りこぼす。
		vOld, okOld := oldVal[key]
		vNew, okNew := newVal[key]

		// 現在のキーのパス (例: "address" -> "address.city")
		currentPath := key
		if path != "" {
			currentPath = path + "." + key
		}

		// キーの有無が異なる場合（追加 or 削除）は、値が nil であっても変更として記録する。
		if okOld != okNew {
			changes = append(changes, ChangePoint{
				Field:    currentPath,
				OldValue: vOld,
				NewValue: vNew,
			})
			continue
		}

		// 値がMap同士なら再帰的に潜る
		mapOld, isMapOld := vOld.(map[string]any)
		mapNew, isMapNew := vNew.(map[string]any)

		if isMapOld && isMapNew {
			// 両方Mapなら再帰呼び出し
			changes = append(changes, compareMaps(currentPath, mapOld, mapNew)...)
			continue
		}

		// 変更があるかチェック
		if !reflect.DeepEqual(vOld, vNew) {
			changes = append(changes, ChangePoint{
				Field:    currentPath,
				OldValue: vOld,
				NewValue: vNew,
			})
		}
	}

	return changes
}
