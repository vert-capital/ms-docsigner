package utils

import "encoding/json"

func Contains[T string | int](elems []T, item T) bool {
	for _, v := range elems {
		if v == item {
			return true
		}
	}
	return false
}

func GenericMapToJson(m map[string]any) []byte {
	json, _ := json.Marshal(m)
	return json
}
