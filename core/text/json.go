package text

import (
	"encoding/json"
)

// SerializeJson は、与えられたデータをJSON文字列にシリアライズします。
// ジェネリック型Tを受け取り、データのJSON文字列表現とエラーを返します。
func SerializeJson[T interface{}](data T) (string, error) {

	result, err := json.Marshal(data)

	if err != nil {
		return "", err
	}

	return string(result), nil
}

// DeserializeJson は、与えられたJSON文字列を指定された型Tにデシリアライズします。
// JSON文字列を受け取り、デシリアライズされた型Tのデータとエラーを返します。
func DeserializeJson[T interface{}](data string) (T, error) {
	var result T

	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
