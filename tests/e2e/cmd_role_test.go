//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"testing"
)

func TestRole(t *testing.T) {
	e := newEnv()

	roleName := makeRoleName()
	roleName1 := roleName + "1"
	roleName2 := roleName + "2"

	output := runWithAuth(t, e, "role")
	requireContains(t, output, "Execute an action on a role")

	output = runWithAuth(t, e, "role --help")
	requireContains(t, output, "Execute an action on a role")

	output = runWithAuth(t, e, "role create --help")
	requireLine(t, output, "Create a role in DevOps Secrets Vault")

	output = runWithAuth(t, e, fmt.Sprintf("role create --name %s --external-id some-id", roleName1))
	requireContains(t, output, "must specify both provider and external ID")

	output = runWithAuth(t, e, fmt.Sprintf("role create --name %s --provider some-provider", roleName1))
	requireContains(t, output, "must specify both provider and external ID")

	output = runWithAuth(t, e, fmt.Sprintf("role create --name %s --desc E2E-CLI-testing", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))
	requireContains(t, output, `"description": "E2E-CLI-testing"`)

	output = runWithAuth(t, e, fmt.Sprintf("role create --name %s --desc E2E-CLI-testing", roleName2))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName2))
	requireContains(t, output, `"description": "E2E-CLI-testing"`)

	output = runWithAuth(t, e, fmt.Sprintf("role %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithAuth(t, e, fmt.Sprintf("role --name %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithAuth(t, e, fmt.Sprintf("role -n %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithAuth(t, e, fmt.Sprintf("role read --name %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithAuth(t, e, fmt.Sprintf("role read -n %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithAuth(t, e, fmt.Sprintf("role delete -n %s", roleName1))
	requireContains(t, output, "role marked for deletion and will be removed in about 72 hours")

	runWithAuth(t, e, fmt.Sprintf("role restore -n %s", roleName1))

	output = runWithAuth(t, e, "role update --desc E2E-CLI-testing-updated")
	requireContains(t, output, "error: must specify name")

	output = runWithAuth(t, e, fmt.Sprintf("role update --name %s --desc E2E-CLI-testing-updated", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))
	requireContains(t, output, `"description": "E2E-CLI-testing-updated"`)
	requireContains(t, output, `"version": "1"`)

	output = runWithAuth(t, e, fmt.Sprintf("role search --query %s --limit 5", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithAuth(t, e, fmt.Sprintf("role search --limit 1 --sort asc --sorted-by name --query %s", roleName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))
	requireNotContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName2))

	output = runWithAuth(t, e, fmt.Sprintf("role search --limit 1 --sort desc --sorted-by name --query %s", roleName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName2))
	requireNotContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithAuth(t, e, "role delete --force")
	requireContains(t, output, "error: must specify name")

	runWithAuth(t, e, fmt.Sprintf("role delete --name %s --force", roleName1))

	output = runWithAuth(t, e, fmt.Sprintf("role delete %s --force", roleName1))
	requireContains(t, output, "unable to find item with specified identifier")

	runWithAuth(t, e, fmt.Sprintf("role delete --name %s --force", roleName2))
}
