package core

import "github.com/golaboratory/gloudia/core/text"

// Clone は与えられたオブジェクトobjをJSONシリアライズとデシリアライズを用いてクローンします。
// オブジェクトの型はT1で指定され、クローンされたオブジェクトの型はT2で指定されます。
// 引数:
//   - obj: クローンする対象のオブジェクト
//
// 戻り値:
//   - T2: クローンされたオブジェクト
//   - error: クローン処理中に発生したエラー
func Clone[T1 any, T2 any](obj T1) (T2, error) {
	var result T2
	json, err := text.SerializeJson(obj)

	if err != nil {
		return result, err
	}

	result, err = text.DeserializeJson[T2](json)
	if err != nil {
		return result, err
	}

	return result, nil
}
