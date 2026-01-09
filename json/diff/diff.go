package diff

import (
	"encoding/json"
	"log/slog"
	"reflect"
	"sort"

	"github.com/newmo-oss/ergo"
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
			return nil, ergo.New("failed to unmarshal old json", slog.String("error", err.Error()))
		}
	}
	if len(newJson) > 0 {
		if err := json.Unmarshal(newJson, &newMap); err != nil {
			return nil, ergo.New("failed to unmarshal new json", slog.String("error", err.Error()))
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
		vOld := oldVal[key]
		vNew := newVal[key]

		// 現在のキーのパス (例: "address" -> "address.city")
		currentPath := key
		if path != "" {
			currentPath = path + "." + key
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
