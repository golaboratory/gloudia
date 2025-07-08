package ref

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// GetFuncName は与えられた関数またはメソッドから、その名称のみを抽出して返します。
// 引数:
//   - v: 関数またはメソッド（interface{}型）
//
// 戻り値:
//   - string: 関数名またはメソッド名（メソッドの場合はパッケージ名や型名を除いた名称）
//   - error: vが関数でない場合やnilの場合にエラーを返します
//
// 備考:
//   - 匿名関数の場合は内部的な名称が返されます
func GetFuncName(v any) (string, error) {
	fv := reflect.ValueOf(v)
	if fv.Kind() != reflect.Func {
		return "", fmt.Errorf("not a function")
	}
	f := runtime.FuncForPC(fv.Pointer())
	if f == nil {
		return "", fmt.Errorf("cannot retrieve function info")
	}
	fullName := strings.TrimSuffix(f.Name(), "-fm")
	if i := strings.LastIndex(fullName, "."); i != -1 {
		return fullName[i+1:], nil
	}
	return fullName, nil
}

// GetStructName は引数として与えられたstructまたはstructのポインターから、structの名称を取得して返します。
// 引数:
//   - s: structまたはstructのポインター
//
// 戻り値:
//   - string: structの名称
//   - error: 引数がstructでない場合にエラーを返します
func GetStructName(s any) (string, error) {
	t := reflect.TypeOf(s)
	if t == nil {
		return "", fmt.Errorf("nil が渡されました")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("struct ではありません")
	}
	return t.Name(), nil
}
