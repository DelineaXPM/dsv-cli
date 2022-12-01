// This contains customized goreleaser tasks that take into account the GOOS and combine this with my standard approach of using changelog to drive the new semver release.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

func checkEnvVar(varName string, tbl pterm.TableData, isSecret bool, notes string) (string, bool, pterm.TableData) {
	var value, valueOfVar string
	var isSet bool

	value, isSet = os.LookupEnv(varName)

	if isSet {
		if isSecret {
			valueOfVar = "***** secret set, but not logged *****"
		} else {
			valueOfVar = value
		}

		tbl = append(tbl, []string{"✅", varName, valueOfVar, notes})
		return value, true, tbl
	}
	tbl = append(tbl, []string{"❌", varName, valueOfVar, notes})
	return "", false, tbl
}

// func checkEnvVar(envVar string, required bool) (string, error) { //nolint:unused // leaving this as will use in future releases
// 	envVarValue := os.Getenv(envVar)
// 	if envVarValue == "" && required {
// 		pterm.Error.Printfln(
// 			"%s is required and unable to proceed without this being provided. terminating task.",
// 			envVar,
// 		)
// 		return "", fmt.Errorf("%s is required", envVar)
// 	}
// 	if envVarValue == "" {
// 		pterm.Debug.Printfln(
// 			"checkEnvVar() found no value for: %q, however this is marked as optional, so not exiting task",
// 			envVar,
// 		)
// 	}
// 	pterm.Debug.Printfln("checkEnvVar() found value: %q=%q", envVar, envVarValue)
// 	return envVarValue, nil
// }

// 🔨 Build builds the project for the current platform.
func Build() error {
	magetoolsutils.CheckPtermDebug()

	configfile, err := getGoreleaserConfig()
	if err != nil {
		return err
	}
	releaserArgs := []string{
		"build",
		"--rm-dist",
		"--snapshot",
		"--single-target",
		"--config", configfile,
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)

	return sh.RunV("goreleaser", releaserArgs...) // "--skip-announce",.
}

// 🔨 BuildAll builds all the binaries defined in the project, for all platforms. This includes Docker image generation but skips publish.
// If there is no additional platforms configured in the task, then basically this will just be the same as `mage build`.
func BuildAll() error {
	magetoolsutils.CheckPtermDebug()

	configfile, err := getGoreleaserConfig()
	if err != nil {
		return err
	}

	releaserArgs := []string{
		"release",
		"--snapshot",
		"--rm-dist",
		"--skip-publish",
		"--config", configfile,
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)
	return sh.RunV("goreleaser", releaserArgs...)
	// To pass in explicit version mapping, you can do this. I'm not using at this time.
	// Return sh.RunWithV(map[string]string{
	// 	"GORELEASER_CURRENT_TAG": "latest",
	// }, binary, releaserArgs...)
}

// 🔨 Release generates a release for the current platform.
func Release() error {
	magetoolsutils.CheckPtermDebug()

	// TODO: this will be checked once we publish cli to github
	// if _, err := checkEnvVar("DOCKER_ORG", true); err != nil {
	// 	return err
	// }

	releaseVersion, err := sh.Output("changie", "latest")
	if err != nil {
		pterm.Error.Printfln("changie pulling latest release note version failure: %v", err)
		return err
	}
	cleanVersion := strings.TrimSpace(releaseVersion)
	cleanpath := filepath.Join(".changes", cleanVersion+".md")
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		cleanpath = filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".changes", cleanVersion+".md")
	}

	configfile, err := getGoreleaserConfig()
	if err != nil {
		return err
	}

	releaserArgs := []string{
		"release",
		"--rm-dist",
		"--skip-validate",
		fmt.Sprintf("--release-notes=%s", cleanpath),
		"--config", configfile,
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)

	return sh.RunWithV(map[string]string{
		"GORELEASER_CURRENT_TAG": cleanVersion,
	},
		"goreleaser",
		releaserArgs...,
	)
}

// getGoreleaserConfig returns the path to the goreleaser config file based on the current OS.
func getGoreleaserConfig() (string, error) {
	magetoolsutils.CheckPtermDebug()

	var configfile string
	switch runtime.GOOS {
	case "darwin":
		configfile = ".goreleaser.darwin.yaml"
	case "linux":
		configfile = ".goreleaser.linux.yaml"
	case "windows":
		configfile = ".goreleaser.windows.yaml"
	default:
		pterm.Error.Printfln("Unsupported OS: %s", runtime.GOOS)
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	pterm.Info.Printfln("using config file: %s", configfile)
	return configfile, nil
}
