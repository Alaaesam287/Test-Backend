package utils

func InterfaceSlice[T any](s []T) []interface{} {
	out := make([]interface{}, len(s))
	for i := range s {
		out[i] = s[i]
	}
	return out
}
