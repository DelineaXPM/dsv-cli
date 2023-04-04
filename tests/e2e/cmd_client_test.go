//go:build endtoend
// +build endtoend

package e2e

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	roleName := makeRoleName()
	output := runWithProfile(t, fmt.Sprintf("role create --name %s --desc E2E-CLI-testing", roleName))
	requireContains(t, output, fmt.Sprintf(`"name": "%s"`, roleName))
	defer func() {
		runWithProfile(t, fmt.Sprintf("role delete --name %s --force", roleName))
	}()

	output = runWithProfile(t, fmt.Sprintf("client create --role %s", roleName))
	requireContains(t, output, `"clientId":`)
	requireContains(t, output, `"clientSecret":`)
	requireContains(t, output, fmt.Sprintf(`"role": "%s"`, roleName))

	// Save client id to delete it later.
	response := make(map[string]any)
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("response is not valid JSON: %v\nResponse:\n%s", err, output)
	}
	clientID := response["clientId"].(string)

	output = runWithProfile(t, fmt.Sprintf("client create --role %s-doesnotexist", roleName))
	requireContains(t, output, `"code": 400`)

	output = runWithProfile(t, fmt.Sprintf("client search --role %s", roleName))
	requireContains(t, output, fmt.Sprintf(`"role": "%s"`, roleName))

	output = runWithProfile(t, fmt.Sprintf("client delete --client-id %s --force", clientID))
	requireEmpty(t, output)
}
