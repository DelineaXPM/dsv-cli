//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"testing"
)

func TestAuthProvider(t *testing.T) {
	authProviderName := makeAuthProviderName()

	output := runWithProfile(t, "config auth-provider")
	requireContains(t, output, "Execute an action on an auth-provider")

	output = runWithProfile(t, "config auth-provider --help")
	requireContains(t, output, "Execute an action on an auth-provider")

	output = runWithProfile(t, fmt.Sprintf(
		"config auth-provider create --name %s --type aws --aws-account-id 1234", authProviderName,
	))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, authProviderName))
	requireContains(t, output, `"type": "aws"`)
	requireContains(t, output, `"properties": {`)
	requireContains(t, output, `"accountId": "1234"`)

	output = runWithProfile(t, fmt.Sprintf(
		"config auth-provider read --name %s", authProviderName,
	))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, authProviderName))
	requireContains(t, output, `"type": "aws"`)
	requireContains(t, output, `"properties": {`)
	requireContains(t, output, `"accountId": "1234"`)

	output = runWithProfile(t, fmt.Sprintf(
		"config auth-provider update --name %s --type aws --aws-account-id 4321", authProviderName,
	))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, authProviderName))
	requireContains(t, output, `"type": "aws"`)
	requireContains(t, output, `"properties": {`)
	requireContains(t, output, `"accountId": "4321"`)

	output = runWithProfile(t, fmt.Sprintf(
		"config auth-provider rollback --name %s", authProviderName,
	))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, authProviderName))
	requireContains(t, output, `"type": "aws"`)
	requireContains(t, output, `"properties": {`)
	requireContains(t, output, `"accountId": "1234"`)

	output = runWithProfile(t, fmt.Sprintf(
		"config auth-provider delete --name %s --force", authProviderName,
	))
	requireEmpty(t, output)
}
