//go:build endtoend
// +build endtoend

package e2e

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestInitWithNoConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyEnter()

		c.ExpectString("Please enter auth type")
		c.SendKeyEnter()

		c.ExpectString("Please enter username")
		c.SendLine(e.username)

		c.ExpectString("Please enter password")
		c.SendLine(e.password)

		c.ExpectEOF()
	})

	got := readFile(t, config)
	requireLine(t, got, "default:")
	requireLine(t, got, "auth:")
	requireLine(t, got, "type: password")
	requireLine(t, got, fmt.Sprintf("username: %s", e.username))
	requireLine(t, got, "cache:")
	requireLine(t, got, "strategy: server")
	requireLine(t, got, fmt.Sprintf("domain: %s", e.domain))
	requireLine(t, got, "store:")
	requireLine(t, got, "type: file")
	requireLine(t, got, fmt.Sprintf("tenant: %s", e.tenant))
}

func TestInitWithExistingConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	const (
		profileName = "automation"
		cacheAge    = "5"
	)
	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	createFile(t, config)
	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Found an existing cli-config located at")
		c.ExpectString("Select an option")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter profile name")
		c.SendLine("a a")

		c.ExpectString("Sorry, your reply was invalid: Profile name contains restricted characters.")
		c.SendLine(profileName)

		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter cache age (minutes until expiration)")
		c.SendLine("a")

		c.ExpectString("Sorry, your reply was invalid: Unable to parse age.")
		c.SendLine("-2")

		c.ExpectString("Sorry, your reply was invalid: Unable to parse age.")
		c.SendLine("0")

		c.ExpectString("Sorry, your reply was invalid: Unable to parse age.")
		c.SendLine(cacheAge)

		c.ExpectString("Please enter auth type")
		c.SendKeyEnter()

		c.ExpectString("Please enter username")
		c.SendLine(e.username)

		c.ExpectString("Please enter password")
		c.SendLine(e.password)

		c.ExpectEOF()
	})

	got := readFile(t, config)
	requireLine(t, got, fmt.Sprintf("%s:", profileName))
	requireLine(t, got, "auth:")
	requireLine(t, got, "type: password")
	requireLine(t, got, fmt.Sprintf("username: %s", e.username))
	requireLine(t, got, "cache:")
	requireLine(t, got, fmt.Sprintf("age: %s", cacheAge))
	requireLine(t, got, "strategy: cache.server")
	requireLine(t, got, fmt.Sprintf("domain: %s", e.domain))
	requireLine(t, got, "store:")
	requireLine(t, got, "type: file")
	requireLine(t, got, fmt.Sprintf("tenant: %s", e.tenant))
}

func TestInitOverwriteExistingConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	createFile(t, config)
	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Found an existing cli-config located at")
		c.ExpectString("Select an option")
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyEnter()

		c.ExpectString("Please enter auth type")
		c.SendKeyEnter()

		c.ExpectString("Please enter username")
		c.SendLine(e.username)

		c.ExpectString("Please enter password")
		c.SendLine(e.password)

		c.ExpectEOF()
	})

	got := readFile(t, config)
	requireLine(t, got, "default:")
	requireLine(t, got, "auth:")
	requireLine(t, got, "type: password")
	requireLine(t, got, fmt.Sprintf("username: %s", e.username))
	requireLine(t, got, "cache:")
	requireLine(t, got, "strategy: server")
	requireLine(t, got, fmt.Sprintf("domain: %s", e.domain))
	requireLine(t, got, "store:")
	requireLine(t, got, "type: file")
	requireLine(t, got, fmt.Sprintf("tenant: %s", e.tenant))
}

func TestInitAuthFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	createFile(t, config)
	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Found an existing cli-config located at")
		c.ExpectString("Select an option")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter profile name")
		c.SendLine("automation")

		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyEnter()

		c.ExpectString("Please enter auth type")
		c.SendKeyEnter()

		c.ExpectString("Please enter username")
		c.SendLine("random-username-that-definitely-does-not-exist")

		c.ExpectString("Please enter password")
		c.SendLine("n0t-a-Stronges-P@assWord")

		c.ExpectString("Failed to authenticate, restoring previous config.")
		c.ExpectEOF()
	})
}

func TestInitAWSInvalid(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	createFile(t, config)
	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Found an existing cli-config located at")
		c.ExpectString("Select an option")
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyEnter()

		c.ExpectString("Please enter auth type")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter aws profile for federated aws auth")
		c.SendLine("not-a-real-profile")

		c.ExpectString("Failed to authenticate, restoring previous config.")
		c.ExpectEOF()
	})
}

func TestInitClientCredsInvalid(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	createFile(t, config)
	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Found an existing cli-config located at")
		c.ExpectString("Select an option")
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyEnter()

		c.ExpectString("Please enter auth type")
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter client id for client auth")
		c.SendLine("not-a-real-client-id")

		c.ExpectString("Please enter client secret for client auth")
		c.SendLine("not-a-real-client-secret")

		c.ExpectString("Failed to authenticate, restoring previous config.")
		c.ExpectEOF()
	})
}

func TestInitMissingInitialProfile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config), "--profile=automation",
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Initial configuration is needed in order to add a custom profile.")
		c.ExpectEOF()
	})
}

func TestInitProfileExists(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyEnter()

		c.ExpectString("Please enter auth type")
		c.SendKeyEnter()

		c.ExpectString("Please enter username")
		c.SendLine(e.username)

		c.ExpectString("Please enter password")
		c.SendLine(e.password)

		c.ExpectEOF()
	})

	cmd = []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config), "--profile=default",
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Profile \"default\" already exists in the config.")
		c.ExpectEOF()
	})
}

func TestInitWrongStoreType(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Failed to get store: 'wincred' option for store.type is supported on windows only.")
		c.ExpectEOF()
	})
}

func TestInitWithNoStore(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter auth type")
		c.SendKeyEnter()

		c.ExpectString("Config created but no credentials saved, specify them as environment variables or via command line flags.")
		c.ExpectEOF()
	})

	got := readFile(t, config)
	requireLine(t, got, "default:")
	requireLine(t, got, "auth:")
	requireLine(t, got, "type: password")
	requireLine(t, got, fmt.Sprintf("domain: %s", e.domain))
	requireLine(t, got, "store:")
	requireLine(t, got, "type: none")
	requireLine(t, got, fmt.Sprintf("tenant: %s", e.tenant))
}

func TestInitUsingCertificateRawData(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	const (
		profileName = "raw-cert-data"
		cacheAge    = "5"
	)
	var (
		config = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
	)

	createFile(t, config)
	defer func() { deleteFile(t, config) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Found an existing cli-config located at")
		c.ExpectString("Select an option")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter profile name")
		c.SendLine(profileName)

		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter cache age (minutes until expiration)")
		c.SendLine(cacheAge)

		c.ExpectString("Please enter auth type")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Raw certificate")
		c.SendKeyEnter()

		c.ExpectString("Raw certificate:")
		time.Sleep(time.Millisecond * 10)
		c.SendLine(e.certificate)

		c.ExpectString("Raw private key")
		c.SendKeyEnter()

		c.ExpectString("Private key:")
		time.Sleep(time.Millisecond * 10)
		c.SendLine(e.privateKey)

		c.ExpectEOF()
	})

	got := readFile(t, config)
	requireLine(t, got, fmt.Sprintf("%s:", profileName))
	requireLine(t, got, "auth:")
	requireLine(t, got, "type: cert")
	requireLine(t, got, fmt.Sprintf("certificate: %s", e.certificate))
	requireLine(t, got, fmt.Sprintf("privateKey: %s", e.privateKey))
	requireLine(t, got, "cache:")
	requireLine(t, got, fmt.Sprintf("age: %s", cacheAge))
	requireLine(t, got, "strategy: cache.server")
	requireLine(t, got, fmt.Sprintf("domain: %s", e.domain))
	requireLine(t, got, "store:")
	requireLine(t, got, "type: file")
	requireLine(t, got, fmt.Sprintf("tenant: %s", e.tenant))
}

func TestInitUsingCertificateFileData(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv(t)

	const (
		profileName = "file-cert-data"
		cacheAge    = "5"
	)
	var (
		config   = filepath.Join(e.tmpDirPath, "e2e-configuration.yml")
		certPath = filepath.Join(e.tmpDirPath, "e2e-cert-data.json")
	)

	certData, err := json.Marshal(&struct {
		Certificate  string `json:"certificate"`
		PrivateKey   string `json:"privateKey"`
		SshPublicKey string `json:"sshPublicKey"`
	}{
		Certificate:  e.certificate,
		PrivateKey:   e.privateKey,
		SshPublicKey: "ssh-key",
	})
	if err != nil {
		t.Fatalf("json.Marshal(%q) = %v", certData, err)
	}

	createFile(t, config)
	defer func() { deleteFile(t, config) }()

	writeFile(t, certData, certPath)
	defer func() { deleteFile(t, certPath) }()

	cmd := []string{
		"init", fmt.Sprintf("--dev=%s", e.domain), fmt.Sprintf("--config=%s", config),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Found an existing cli-config located at")
		c.ExpectString("Select an option")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter profile name")
		c.SendLine(profileName)

		c.ExpectString("Please enter tenant name")
		c.SendLine(e.tenant)

		c.ExpectString("Please select store type")
		c.SendKeyEnter()

		c.ExpectString("Please enter directory for file store")
		c.SendKeyEnter()

		c.ExpectString("Please enter cache strategy for secrets")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Please enter cache age (minutes until expiration)")
		c.SendLine(cacheAge)

		c.ExpectString("Please enter auth type")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Raw certificate")
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Certificate file path:")
		time.Sleep(time.Millisecond * 10)
		c.SendLine(certPath)

		c.ExpectString("Raw private key")
		c.SendKeyArrowDown()
		c.SendKeyEnter()

		c.ExpectString("Private key file path:")
		time.Sleep(time.Millisecond * 10)
		c.SendLine(certPath)

		c.ExpectEOF()
	})

	got := readFile(t, config)
	requireLine(t, got, fmt.Sprintf("%s:", profileName))
	requireLine(t, got, "auth:")
	requireLine(t, got, "type: cert")
	requireLine(t, got, fmt.Sprintf("certificate: %s", e.certificate))
	requireLine(t, got, fmt.Sprintf("privateKey: %s", e.privateKey))
	requireLine(t, got, "cache:")
	requireLine(t, got, fmt.Sprintf("age: %s", cacheAge))
	requireLine(t, got, "strategy: cache.server")
	requireLine(t, got, fmt.Sprintf("domain: %s", e.domain))
	requireLine(t, got, "store:")
	requireLine(t, got, "type: file")
	requireLine(t, got, fmt.Sprintf("tenant: %s", e.tenant))
}
