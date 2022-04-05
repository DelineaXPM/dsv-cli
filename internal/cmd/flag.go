package cmd

import (
	"strings"
)

func FriendlyName(flag string) string {
	return strings.Replace(flag, ".", "-", -1)
}
