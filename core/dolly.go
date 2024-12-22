package core

import "golaboratory/gloudia/core/text"

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
