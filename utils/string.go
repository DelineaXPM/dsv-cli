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

//CheckPrefix check prefixes and returns true if one of prefixes satisfies the condition
func CheckPrefix(path string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// EqAny checks whether string equals any within a list of candidates
func EqAny(str string, candidates []string) bool {
	for _, c := range candidates {
		if c == str {
			return true
		}
	}

	return false
}

func ToPointerString(s string) *string {
	return &s
}
