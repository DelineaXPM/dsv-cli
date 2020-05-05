package utils

import "strings"

// StringToSlice converts a string with comma-separated elements to a slice.
// It first excludes a leading or trailing bracket character if present in the string.
func StringToSlice(str string) []string {
	if str == "" {
		return []string{""}
	}
	if str[0] == '[' {
		str = str[1:]
	}
	length := len(str)
	if str[length-1] == ']' {
		str = str[:length-1]
	}
	return strings.Split(str, ",")
}
