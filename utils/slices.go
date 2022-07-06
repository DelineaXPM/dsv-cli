package utils

func IndexOf(slice []string, target string) int {
	for i, ele := range slice {
		if ele == target {
			return i
		}
	}
	return -1
}
