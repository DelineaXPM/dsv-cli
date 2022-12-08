package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/DelineaXPM/dsv-cli/magefiles/constants"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
)

// &mage go:testsum ./tests/e2e/...
// displayName: mage test
// workingDirectory: ${{ parameters.workingDirectory }}/$(Build.Repository.Name)
// failOnStderr: false
// env:
//   AQUA_ROOT_DIR: $(AQUA_ROOT_DIR)
//   GOROOT: $(goenv.GOROOT)
//   GOPATH: $(GOPATH)
//   GO_INTEGRATION_TEST: 1 # REQUIRED TO ALLOW INTEGRATION TEST TO TRIGGER
//   GOTEST_DISABLE_RACE: 1
//   GOTEST_FLAGS: '-tags=endtoend'

//   # secrets that have to be exposed to script to be visible
//   ADMIN_USER: $(ADMIN_USER)
//   DSV_ADMIN_PASS: $(DSV_ADMIN_PASS)
//   DSV_USER_PASSWORD: $(DSV_USER_PASSWORD)
//   CLIENT_ID: $(CLIENT_ID)
//   DSV_CLIENT_SECRET: $(DSV_CLIENT_SECRET)

//   ## Other config values just for clarity
//   TEST_TENANT: $(TEST_TENANT)
//   USER_NAME: $(USER_NAME)
//   LOCAL_DOMAIN: $(LOCAL_DOMAIN)
//   DEV_DOMAIN: $(DEV_DOMAIN)

// AQUA_ROOT_DIR: $(AQUA_ROOT_DIR)
// GOROOT: $(goenv.GOROOT)
// GOPATH: $(GOPATH)
// GO_INTEGRATION_TEST: 1 # REQUIRED TO ALLOW INTEGRATION TEST TO TRIGGER
// GOTEST_DISABLE_RACE: 1

// # secrets that have to be exposed to script to be visible
// ADMIN_USER: $(ADMIN_USER)
// DSV_ADMIN_PASS: $(DSV_ADMIN_PASS)
// DSV_USER_PASSWORD: $(DSV_USER_PASSWORD)
// CLIENT_ID: $(CLIENT_ID)
// DSV_CLIENT_SECRET: $(DSV_CLIENT_SECRET)

// ## Other config values just for clarity
// TEST_TENANT: $(TEST_TENANT)
// USER_NAME: $(USER_NAME)
// LOCAL_DOMAIN: $(LOCAL_DOMAIN)
// DEV_DOMAIN: $(DEV_DOMAIN)
type Test mg.Namespace

func (Test) Integration() error {
	var err error
	// track total failing conditions and report back.
	var errorCount int
	tbl := pterm.TableData{
		[]string{"Status", "Check", "Value", "Notes"},
		[]string{"✅", "GOVERSION", runtime.Version(), ""},
		[]string{"✅", "GOOS", runtime.GOOS, ""},
		[]string{"✅", "GOARCH", runtime.GOARCH, ""},
		[]string{"✅", "GOROOT", runtime.GOROOT(), ""},
		[]string{"✅", "GOPATH", os.Getenv("GOPATH"), ""},
	}
	defer func(tbl *pterm.TableData) {
		primary := pterm.NewStyle(pterm.FgLightWhite, pterm.BgGray, pterm.Bold)

		if err := pterm.DefaultTable.WithHasHeader().
			WithBoxed(true).
			WithHeaderStyle(primary).
			WithData(*tbl).Render(); err != nil {
			pterm.Error.Printf(
				"pterm.DefaultTable.WithHasHeader of variable information failed. Continuing...\n%v",
				err,
			)
		}
	}(&tbl)

	_, tbl, err = checkEnvVar(checkEnv{Name: "GO_INTEGRATION_TEST", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "this must be set to 1 for the integration test to run"})
	if err != nil {
		errorCount++
	}

	// not really a secret, but not necessary to log this user name so setting as secret to hide output.
	_, tbl, err = checkEnvVar(checkEnv{Name: "ADMIN_USER", IsSecret: true, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "DSV_ADMIN_PASS", IsSecret: true, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "DSV_USER_PASSWORD", IsSecret: true, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "CLIENT_ID", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "DSV_CLIENT_SECRET", IsSecret: true, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "TEST_TENANT", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "USER_NAME", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	// Not used in this repos test, maybe robot?
	// _, tbl, err = checkEnvVar(checkEnv{Name: "USER_NAME1", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	// if err != nil {
	// 	errorCount++
	// }
	_, tbl, err = checkEnvVar(checkEnv{Name: "LOCAL_DOMAIN", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "DEV_DOMAIN", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required environment variable"})
	if err != nil {
		errorCount++
	}

	if errorCount > 0 {
		pterm.Error.Printfln("terminating task since errorCount '%d' exceeds 0", errorCount)
		return fmt.Errorf("terminating task since errorCount '%d' exceeds 0", errorCount)
	}

	return sh.RunWithV(map[string]string{
		"GO_INTEGRATION_TEST": "1",
	},
		"gotestsum",
		"--format", "pkgname",
		// normal test args go after the dash dash. Args before this are for gotestsum.
		"--",
		"-shuffle=on",
		constants.IntegrationDirectoryTestPath,
	)
}
