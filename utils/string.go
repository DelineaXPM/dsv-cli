package utils

import (
	"strings"
)

// StringToSlice converts a string with comma-separated elements to a slice.
// It first excludes a leading or trailing bracket character if present in the string.
func StringToSlice(str string) []string {
	if str == "" {
		return []string{""}
	}
	str = strings.TrimPrefix(str, "[")
	str = strings.TrimSuffix(str, "]")
	str = strings.ReplaceAll(str, " ", "")
	return strings.Split(str, ",")
}
