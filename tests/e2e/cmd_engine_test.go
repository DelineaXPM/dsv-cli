//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"testing"
)

func TestEngine(t *testing.T) {
	e := newEnv(t)

	poolName := makePoolName()
	engineName := makeEngineName()

	output := runWithAuth(t, e, fmt.Sprintf("pool create --name=%s", poolName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("engine read --name=%s", engineName))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)

	output = runWithAuth(t, e, fmt.Sprintf("engine create --name=%s --pool-name=e2e-cli-test-pool-that-does-not-exist", engineName))
	requireContains(t, output, `"message": "specified pool doesn't exist"`)

	output = runWithAuth(t, e, fmt.Sprintf("engine create --name=%s --pool-name=%s", engineName, poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("engine read --name=%s", engineName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, "engine list")
	requireContains(t, output, `"engines": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithAuth(t, e, fmt.Sprintf("engine delete --name=%s", engineName))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}

	output = runWithAuth(t, e, fmt.Sprintf("engine read --name=%s", engineName))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)

	runWithAuth(t, e, fmt.Sprintf("pool delete --name=%s", poolName))
}
