//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"runtime"
	"testing"
)

func TestPool(t *testing.T) {
	e := newEnv()

	poolName1 := makePoolName()
	poolName2 := makePoolName()

	output := runWithProfile(t, "pool")
	requireLine(t, output, "Work with engine pools")

	output = runWithProfile(t, "pool --help")
	requireLine(t, output, "Work with engine pools")

	output = runWithProfile(t, "pool create --help")
	requireContains(t, output, "Create a new empty pool of engines")

	output = runWithProfile(t, fmt.Sprintf("pool create --name=%s", poolName1))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName1))

	output = runWithProfile(t, "pool read --help")
	requireContains(t, output, "Get information on an existing pool of engines")

	output = runWithProfile(t, "pool read")
	requireContains(t, output, "error: must specify name")

	output = runWithProfile(t, fmt.Sprintf("pool read --name=%s", poolName1))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName1))

	output = runWithProfile(t, fmt.Sprintf("pool read %s", poolName1))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName1))

	output = runWithProfile(t, fmt.Sprintf("pool %s", poolName1))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName1))

	output = runWithProfile(t, fmt.Sprintf("pool create --name=%s", poolName2))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName2))

	output = runWithProfile(t, "pool list --help")
	requireContains(t, output, "List the names of all existing pools")

	output = runWithProfile(t, "pool list")
	requireContains(t, output, `"pools": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName2))

	output = runWithProfile(t, "pool list --limit 1")
	requireContains(t, output, `"length": 1`)
	requireContains(t, output, `"limit": 1`)
	requireContains(t, output, `"pools": [`)

	output = runWithProfile(t, "pool delete --help")
	requireContains(t, output, "Delete an existing pool of engines")

	output = runWithProfile(t, "pool delete")
	requireContains(t, output, "error: must specify name")

	output = runWithProfile(t, fmt.Sprintf("pool delete --name=%s", poolName1))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}
	output = runWithProfile(t, fmt.Sprintf("pool delete --name=%s", poolName2))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}
	output = runWithProfile(t, fmt.Sprintf("pool delete %s", poolName1))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)
}

func TestPoolInteractiveCreate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv()

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
	output := runWithProfile(t, fmt.Sprintf("pool delete --name=%s", poolName))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}
}
