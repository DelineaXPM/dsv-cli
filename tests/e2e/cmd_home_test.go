//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"testing"
)

func TestHome(t *testing.T) {
	username := newEnv().username
	path := makeHomeSecretPath()

	output := runWithProfile(t, "home")
	requireContains(t, output, "Work with secrets in a personal user space")

	output = runWithProfile(t, fmt.Sprintf("home create %s --desc zero-description-0", path))
	requireContains(t, output, `"description": "zero-description-0"`)
	requireContains(t, output, fmt.Sprintf(`"path": "users:%s:%s"`, username, path))

	output = runWithProfile(t, fmt.Sprintf("home update %s --desc one-description-1", path))
	requireContains(t, output, `"description": "one-description-1"`)
	requireContains(t, output, `"version": "1"`)
	requireContains(t, output, fmt.Sprintf(`"path": "users:%s:%s"`, username, path))

	output = runWithProfile(t, fmt.Sprintf("home read %s", path))
	requireContains(t, output, `"description": "one-description-1"`)
	requireContains(t, output, `"version": "1"`)
	requireContains(t, output, fmt.Sprintf(`"path": "users:%s:%s"`, username, path))

	output = runWithProfile(t, fmt.Sprintf("home delete %s", path))
	requireContains(t, output, "marked for deletion")

	runWithProfile(t, fmt.Sprintf("home restore %s", path))

	output = runWithProfile(t, fmt.Sprintf("home update %s --desc two-description-2", path))
	requireContains(t, output, `"description": "two-description-2"`)
	requireContains(t, output, `"version": "2"`)
	requireContains(t, output, fmt.Sprintf(`"path": "users:%s:%s"`, username, path))

	output = runWithProfile(t, fmt.Sprintf("home rollback %s --version 0", path))
	requireContains(t, output, `"description": "zero-description-0"`)
	requireContains(t, output, `"version": "3"`)
	requireContains(t, output, fmt.Sprintf(`"path": "users:%s:%s"`, username, path))

	output = runWithProfile(t, fmt.Sprintf("home read %s", path))
	requireContains(t, output, `"description": "zero-description-0"`)
	requireContains(t, output, `"version": "3"`)
	requireContains(t, output, fmt.Sprintf(`"path": "users:%s:%s"`, username, path))

	output = runWithProfile(t, fmt.Sprintf("home read --path %s --version 3", path))
	requireContains(t, output, `"description": "two-description-2"`)
	requireContains(t, output, `"description": "one-description-1"`)
	requireContains(t, output, `"description": "zero-description-0"`)
	requireContains(t, output, fmt.Sprintf(`"path": "users:%s:%s"`, username, path))

	runWithProfile(t, fmt.Sprintf("home delete %s --force", path))
}
