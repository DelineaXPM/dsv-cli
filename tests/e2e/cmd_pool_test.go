//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"runtime"
	"testing"
)

func TestPool(t *testing.T) {
	e := newEnv(t)

	poolName := makePoolName()

	output := runWithAuth(t, e, "pool")
	requireLine(t, output, "Work with engine pools")

	output = runWithAuth(t, e, "pool --help")
	requireLine(t, output, "Work with engine pools")

	output = runWithAuth(t, e, "pool create --help")
	requireContains(t, output, "Create a new empty pool of engines")

	output = runWithAuth(t, e, fmt.Sprintf("pool create --name=%s", poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithAuth(t, e, "pool read --help")
	requireContains(t, output, "Get information on an existing pool of engines")

	output = runWithAuth(t, e, "pool read")
	requireContains(t, output, "error: must specify name")

	output = runWithAuth(t, e, fmt.Sprintf("pool read --name=%s", poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("pool read %s", poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("pool %s", poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithAuth(t, e, "pool list --help")
	requireContains(t, output, "List the names of all existing pools")

	output = runWithAuth(t, e, "pool list")
	requireContains(t, output, `"pools": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithAuth(t, e, "pool delete --help")
	requireContains(t, output, "Delete an existing pool of engines")

	output = runWithAuth(t, e, "pool delete")
	requireContains(t, output, "error: must specify name")

	output = runWithAuth(t, e, fmt.Sprintf("pool delete --name=%s", poolName))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}
	output = runWithAuth(t, e, fmt.Sprintf("pool delete %s", poolName))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)
}

func TestPoolInteractiveCreate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	poolName := makePoolName()

	cmd := []string{
		"pool", "create",
		"--auth-type=password",
		fmt.Sprintf("--auth-username=%s", e.username),
		fmt.Sprintf("--auth-password=%s", e.password),
		fmt.Sprintf("--tenant=%s", e.tenant),
		fmt.Sprintf("--domain=%s", e.domain),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Pool name")
		c.SendLine(poolName)
		c.ExpectEOF()
	})
	output := runWithAuth(t, e, fmt.Sprintf("pool delete --name=%s", poolName))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}
}
