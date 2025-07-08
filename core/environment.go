package core

import (
	"os"
	"strconv"
)

// GetStringVariable は環境変数から文字列の値を取得します。
// 指定されたキーが存在しない場合はデフォルト値を返します。
// 引数:
//   - key: 取得する環境変数のキー
//   - defaultValue: キーが存在しない場合のデフォルト値
//
// 戻り値:
//   - string: 取得した値またはデフォルト値
func GetStringVariable(key string, defaultValue string) string {
	result, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return result
}

// GetIntVariable は環境変数から整数の値を取得します。
// 指定されたキーが存在しないか、値が整数に変換できない場合はデフォルト値を返します。
// 引数:
//   - key: 取得する環境変数のキー
//   - defaultValue: キーが存在しない場合のデフォルト値
//
// 戻り値:
//   - int: 取得した値またはデフォルト値
func GetIntVariable(key string, defaultValue int) int {
	tmp, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	result, err := strconv.Atoi(tmp)
	if err != nil {
		return defaultValue
	}

	return result
}

// GetBoolVariable は環境変数からブール値を取得します。
// 指定されたキーが存在しない場合はデフォルト値を返します。
// 値が "1" の場合は true、それ以外の場合は false を返します。
// 引数:
//   - key: 取得する環境変数のキー
//   - defaultValue: キーが存在しない場合のデフォルト値
//
// 戻り値:
//   - bool: 取得した値またはデフォルト値
func GetBoolVariable(key string, defaultValue bool) bool {
	tmp, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	result := tmp == "1"
	return result
}
