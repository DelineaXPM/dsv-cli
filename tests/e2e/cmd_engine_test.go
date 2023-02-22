//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"sort"
	"testing"
)

func TestEngine(t *testing.T) {
	e := newEnv()

	poolName := makePoolName()
	engineName1 := makeEngineName()
	engineName2 := makeEngineName()
	engineNamesInOrder := []string{engineName1, engineName2}
	sort.Strings(engineNamesInOrder)

	output := runWithProfile(t, fmt.Sprintf("pool create --name=%s", poolName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, poolName))

	output = runWithProfile(t, fmt.Sprintf("engine read --name=%s", engineName1))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)

	output = runWithProfile(t, fmt.Sprintf("engine create --name=%s --pool-name=e2e-cli-test-pool-that-does-not-exist", engineName1))
	requireContains(t, output, `"message": "specified pool doesn't exist"`)

	output = runWithProfile(t, fmt.Sprintf("engine create --name=%s --pool-name=%s", engineName1, poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithProfile(t, fmt.Sprintf("engine create --name=%s --pool-name=%s", engineName2, poolName))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName2))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithProfile(t, fmt.Sprintf("engine read --name=%s", engineName1))
	requireContains(t, output, fmt.Sprintf(`"createdBy": "users:%s",`, e.username))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithProfile(t, "engine list")
	requireContains(t, output, `"engines": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName2))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithProfile(t, "engine list --sort asc --sorted-by name --pool-name "+poolName)
	requireContains(t, output, `"engines": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineNamesInOrder[0]))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineNamesInOrder[1]))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithProfile(t, "engine list --query "+engineName1)
	requireContains(t, output, `"engines": [`)
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName1))
	requireNotContains(t, output, fmt.Sprintf(`"name": "%s"`, engineName2))
	requireContains(t, output, fmt.Sprintf(`"poolName": "%s"`, poolName))

	output = runWithProfile(t, fmt.Sprintf("engine delete --name=%s", engineName1))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}

	output = runWithProfile(t, fmt.Sprintf("engine delete --name=%s", engineName2))
	if output != "" {
		t.Fatalf("Unexpected output: \n%s\n", output)
	}

	output = runWithProfile(t, fmt.Sprintf("engine read --name=%s", engineName1))
	requireContains(t, output, `"message": "unable to find item with specified identifier"`)

	runWithProfile(t, fmt.Sprintf("pool delete --name=%s", poolName))
}
