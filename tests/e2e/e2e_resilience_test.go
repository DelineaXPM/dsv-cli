//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	commonPrefix = "e2e-cli-test"

	// List of keys for `usedObjects` map.
	authProvidersKey = "auth-providers"
	rolesKey         = "roles"
	poolsKey         = "pools"
	enginesKey       = "engines"
	siemKey          = "siem"
	homeKey          = "home"
)

var (
	muxUsedObjects sync.Mutex
	usedObjects    = map[string][]string{}
)

func makeAuthProviderName() string { return makeName(authProvidersKey) }
func makeRoleName() string         { return makeName(rolesKey) }
func makePoolName() string         { return makeName(poolsKey) }
func makeEngineName() string       { return makeName(enginesKey) }
func makeSIEMName() string         { return makeName(siemKey) }

func makeName(key string) string {
	name := fmt.Sprintf("%s-%s-%d", commonPrefix, key, time.Now().UnixNano())

	muxUsedObjects.Lock()
	if _, ok := usedObjects[key]; !ok {
		usedObjects[key] = []string{name}
	} else {
		usedObjects[key] = append(usedObjects[key], name)
	}
	muxUsedObjects.Unlock()

	return name
}

func makeHomeSecretPath() string {
	name := fmt.Sprintf("%s:%s-%d", commonPrefix, homeKey, time.Now().UnixNano())

	muxUsedObjects.Lock()
	if _, ok := usedObjects[homeKey]; !ok {
		usedObjects[homeKey] = []string{name}
	} else {
		usedObjects[homeKey] = append(usedObjects[homeKey], name)
	}
	muxUsedObjects.Unlock()

	return name
}

func resilienceBefore() error {
	// TODO: implement. This function should help to keep tenant clean.
	// All previously created by this testing suite objects should be deleted.
	fmt.Fprintln(os.Stderr, "[ResilienceBefore] <not implemented>")
	return nil
}

func resilienceAfter() error {
	muxUsedObjects.Lock()
	defer muxUsedObjects.Unlock()

	if len(usedObjects) == 0 {
		fmt.Fprintln(os.Stderr, "[ResilienceAfter] Used objects list is empty. Nothing to cleanup.")
		return nil
	}

	if len(usedObjects[authProvidersKey]) > 0 {
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] Recorded %d used auth provider(s).\n", len(usedObjects[authProvidersKey]))
		for _, name := range usedObjects[authProvidersKey] {
			delete(fmt.Sprintf("config auth-provider delete %s --force", name))
		}
	}

	if len(usedObjects[rolesKey]) > 0 {
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] Recorded %d used role(s).\n", len(usedObjects[rolesKey]))
		for _, name := range usedObjects[rolesKey] {
			delete(fmt.Sprintf("role delete %s --force", name))
		}
	}

	if len(usedObjects[enginesKey]) > 0 {
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] Recorded %d used engine(s).\n", len(usedObjects[enginesKey]))
		for _, name := range usedObjects[enginesKey] {
			delete(fmt.Sprintf("engine delete %s", name))
		}
	}

	if len(usedObjects[poolsKey]) > 0 {
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] Recorded %d used pool(s).\n", len(usedObjects[poolsKey]))
		for _, name := range usedObjects[poolsKey] {
			delete(fmt.Sprintf("pool delete %s", name))
		}
	}

	if len(usedObjects[siemKey]) > 0 {
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] Recorded %d used siem(s).\n", len(usedObjects[siemKey]))
		for _, name := range usedObjects[siemKey] {
			delete(fmt.Sprintf("siem delete %s", name))
		}
	}

	if len(usedObjects[homeKey]) > 0 {
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] Recorded %d used Home Vault secret(s).\n", len(usedObjects[homeKey]))
		for _, name := range usedObjects[homeKey] {
			delete(fmt.Sprintf("home delete %s --force", name))
		}
	}

	return nil
}

func delete(command string) {
	e := newEnv()

	binArgs := append(strings.Split(command, " "),
		"--auth-type=password",
		fmt.Sprintf("--auth-username=%s", e.username),
		fmt.Sprintf("--auth-password=%s", e.password),
		fmt.Sprintf("--tenant=%s", e.tenant),
		fmt.Sprintf("--domain=%s", e.domain),
	)

	cmd := exec.Command(binPath, binArgs...)
	cmd.Env = append(os.Environ(), "IS_SYSTEM_TEST=true")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] dsv %s => %v\n", command, err)
		return
	}

	out := string(output)
	// Remove lines like this added at the end of the output:
	// 		> PASS
	// 		> coverage: 6.8% of statements in ./...
	// 		>
	out = out[:strings.Index(out, `PASS`)]

	switch {
	case strings.Contains(out, "unable to find item with specified identifier"):
		// ignore. Obj was removed.

	case strings.Contains(out, "Invalid permissions"):
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] dsv %s => Invalid permissions\n", command)

	case out != "":
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] dsv %s => %s\n", command, out)

	case out == "":
		fmt.Fprintf(os.Stderr, "[ResilienceAfter] dsv %s => success\n", command)
	}
}
