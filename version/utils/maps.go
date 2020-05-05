package utils

func Keys(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
