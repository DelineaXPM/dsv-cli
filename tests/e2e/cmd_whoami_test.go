//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"strings"
	"testing"
)

func TestWhoami(t *testing.T) {
	e := newEnv(t)

	cmd := []string{
		"whoami",
		"--auth-type=password",
		fmt.Sprintf("--auth-username=%s", e.username),
		fmt.Sprintf("--auth-password=%s", e.password),
		fmt.Sprintf("--tenant=%s", e.tenant),
		fmt.Sprintf("--domain=%s", e.domain),
	}
	output := run(t, cmd)
	if !strings.Contains(output, fmt.Sprintf("users:%s", e.username)) {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}
}
