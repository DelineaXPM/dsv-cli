//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"testing"
)

func TestRole(t *testing.T) {
	roleName := makeRoleName()
	roleName1 := roleName + "1"
	roleName2 := roleName + "2"

	output := runWithProfile(t, "role")
	requireContains(t, output, "Execute an action on a role")

	output = runWithProfile(t, "role --help")
	requireContains(t, output, "Execute an action on a role")

	output = runWithProfile(t, "role create --help")
	requireLine(t, output, "Create a role in DevOps Secrets Vault")

	output = runWithProfile(t, fmt.Sprintf("role create --name %s --external-id some-id", roleName1))
	requireContains(t, output, "must specify both provider and external ID")

	output = runWithProfile(t, fmt.Sprintf("role create --name %s --provider some-provider", roleName1))
	requireContains(t, output, "must specify both provider and external ID")

	output = runWithProfile(t, fmt.Sprintf("role create --name %s --desc E2E-CLI-testing", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))
	requireContains(t, output, `"description": "E2E-CLI-testing"`)

	output = runWithProfile(t, fmt.Sprintf("role create --name %s --desc E2E-CLI-testing", roleName2))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName2))
	requireContains(t, output, `"description": "E2E-CLI-testing"`)

	output = runWithProfile(t, fmt.Sprintf("role %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithProfile(t, fmt.Sprintf("role --name %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithProfile(t, fmt.Sprintf("role -n %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithProfile(t, fmt.Sprintf("role read --name %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithProfile(t, fmt.Sprintf("role read -n %s", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithProfile(t, fmt.Sprintf("role delete -n %s", roleName1))
	requireContains(t, output, "role marked for deletion and will be removed in about 72 hours")

	runWithProfile(t, fmt.Sprintf("role restore -n %s", roleName1))

	output = runWithProfile(t, "role update --desc E2E-CLI-testing-updated")
	requireContains(t, output, "error: must specify name")

	output = runWithProfile(t, fmt.Sprintf("role update --name %s --desc E2E-CLI-testing-updated", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))
	requireContains(t, output, `"description": "E2E-CLI-testing-updated"`)
	requireContains(t, output, `"version": "1"`)

	output = runWithProfile(t, fmt.Sprintf("role search --query %s --limit 5", roleName1))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithProfile(t, fmt.Sprintf("role search --limit 1 --sort asc --sorted-by name --query %s", roleName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))
	requireNotContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName2))

	output = runWithProfile(t, fmt.Sprintf("role search --limit 1 --sort desc --sorted-by name --query %s", roleName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName2))
	requireNotContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName1))

	output = runWithProfile(t, "role delete --force")
	requireContains(t, output, "error: must specify name")

	runWithProfile(t, fmt.Sprintf("role delete --name %s --force", roleName1))

	output = runWithProfile(t, fmt.Sprintf("role delete %s --force", roleName1))
	requireContains(t, output, "unable to find item with specified identifier")

	runWithProfile(t, fmt.Sprintf("role delete --name %s --force", roleName2))
}
