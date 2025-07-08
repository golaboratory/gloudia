package text

import (
	"encoding/json"
)

// SerializeJson は与えられたデータをJSON文字列にシリアライズします。
// 引数:
//   - data: シリアライズ対象のデータ（ジェネリック型T）
//
// 戻り値:
//   - string: JSON文字列表現
//   - error: シリアライズ時のエラー
func SerializeJson[T interface{}](data T) (string, error) {

	result, err := json.Marshal(data)

	if err != nil {
		return "", err
	}

	return string(result), nil
}

// DeserializeJson は与えられたJSON文字列を指定された型Tにデシリアライズします。
// 引数:
//   - data: デシリアライズ対象のJSON文字列
//
// 戻り値:
//   - T: デシリアライズされたデータ
//   - error: デシリアライズ時のエラー
func DeserializeJson[T interface{}](data string) (T, error) {
	var result T

	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
