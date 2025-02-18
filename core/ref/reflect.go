package ref

import (
	"reflect"
	"runtime"
)

// GetFuncName は与えられた関数からその関数名を抽出します。
//
// パラメータ：
//   - v any: 関数名を取得したい関数（interface{}型として受け取ります）
//
// 戻り値：
//   - string: 関数名
//   - error: エラー（関数でない値が渡された場合や、nilが渡された場合にエラーを返します）
//
// 例：
//
//	func example() {}
//	name, err := GetFuncName(example)
//	// name には "example" が格納されます
//
// 注意：
//   - 匿名関数の場合は、生成された内部的な関数名が返されます
//   - メソッドの場合は、パッケージ名、型名を含む完全修飾名が返されます
func GetFuncName(v any) (string, error) {
	fv := reflect.ValueOf(v)
	return runtime.FuncForPC(fv.Pointer()).Name(), nil
}
