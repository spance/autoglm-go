package utils

func AnyToString(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func AnyToIntSlice(v any) []int {
	s, ok := v.([]int)
	if !ok {
		return []int{}
	}
	return s
}
