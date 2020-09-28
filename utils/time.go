package utils

import (
	"errors"
	"strconv"
)

// ParseHours returns the number of hours given a string that can represent either an integer (denoting hours)
// or an integer with a time-unit specifier.
func ParseHours(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	var invalidFormatErr = errors.New("invalid time format - submit an integer followed by one of the following: h, H, d, D, w, W")
	units := []string{"h", "H", "d", "D", "w", "W"}
	n, unit := s[:len(s)-1], string(s[len(s)-1])

	// If no valid time unit is specified, assume s is an integer denoting hours.
	if !Contains(units, unit) {
		num, err := strconv.Atoi(s)
		if err != nil {
			return 0, invalidFormatErr
		}
		return num, nil
	}

	// If the last character of s is a valid time unit, ensure the leading part of s can be parsed as an integer.
	num, err := strconv.Atoi(n)
	if err != nil {
		return 0, invalidFormatErr
	}

	switch unit {
	case "h", "H":
		return num, nil
	case "d", "D":
		return num * 24, nil
	case "w", "W":
		return num * 24 * 7, nil
	default:
		return 0, invalidFormatErr
	}
}
