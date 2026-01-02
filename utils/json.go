package utils

import (
	json "github.com/bytedance/sonic"
)

func JsonString(obj any) string {
	jsonStr, _ := json.Marshal(obj)
	return string(jsonStr)
}

func JsonIndent(obj any) string {
	jsonStr, _ := json.MarshalIndent(obj, "", "  ")
	return string(jsonStr)
}
