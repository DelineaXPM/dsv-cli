package utils

// SlicesEqual returns true if two byte slices are equal. Returns false otherwise.
func SlicesEqual(a, b []byte) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Contains(slice []string, target string) bool {
	for _, ele := range slice {
		if ele == target {
			return true
		}
	}
	return false
}
