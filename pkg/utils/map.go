package utils

func CopyStringMap[T any](m map[string]T) map[string]T {
	result := make(map[string]T, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
