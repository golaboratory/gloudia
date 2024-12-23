package core

import (
	"os"
	"strconv"
)

// GetStringVariable は、環境変数から文字列の値を取得します。
// 指定されたキーが存在しない場合は、デフォルト値を返します。
func GetStringVariable(key string, defaultValue string) string {
	result, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return result
}

// GetIntVariable は、環境変数から整数の値を取得します。
// 指定されたキーが存在しないか、値が整数に変換できない場合は、デフォルト値を返します。
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

// GetBoolVariable は、環境変数からブール値を取得します。
// 指定されたキーが存在しない場合は、デフォルト値を返します。
// 値が "1" の場合は true を、それ以外の場合は false を返します。
func GetBoolVariable(key string, defaultValue bool) bool {
	tmp, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	result := tmp == "1"
	return result
}
