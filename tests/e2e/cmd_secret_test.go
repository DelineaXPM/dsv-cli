//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"strings"
	"testing"
)

func TestSecret(t *testing.T) {
	e := newEnv()

	cmd := []string{
		"secret",
		"--auth-type=password",
		fmt.Sprintf("--auth-username=%s", e.username),
		fmt.Sprintf("--auth-password=%s", e.password),
		fmt.Sprintf("--tenant=%s", e.tenant),
		fmt.Sprintf("--domain=%s", e.domain),
	}
	output := run(t, cmd)
	if !strings.Contains(output, "Execute an action on a secret") {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}
}
