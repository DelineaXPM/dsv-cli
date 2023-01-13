//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"testing"
)

func TestEngine(t *testing.T) {
	e := newEnv()

	poolName := makePoolName()
	engineName1 := makeEngineName()
	engineName2 := makeEngineName()

	output := runWithAuth(t, e, fmt.Sprintf("pool create --name=%s", poolName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("engine read --name=%s", engineName1))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)

	output = runWithAuth(t, e, fmt.Sprintf("engine create --name=%s --pool-name=e2e-cli-test-pool-that-does-not-exist", engineName1))
	requireContains(t, output, `"message": "specified pool doesn't exist"`)

	output = runWithAuth(t, e, fmt.Sprintf("engine create --name=%s --pool-name=%s", engineName1, poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("engine create --name=%s --pool-name=%s", engineName2, poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName2))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("engine read --name=%s", engineName1))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, "engine list")
	requireContains(t, output, `"engines": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName2))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, "engine list --limit 1")
	requireContains(t, output, `"engines": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireNotContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName2))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("engine delete --name=%s", engineName1))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}

	output = runWithAuth(t, e, fmt.Sprintf("engine delete --name=%s", engineName2))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}

	output = runWithAuth(t, e, fmt.Sprintf("engine read --name=%s", engineName1))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)

	runWithAuth(t, e, fmt.Sprintf("pool delete --name=%s", poolName))
}
