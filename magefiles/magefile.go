// 🧙 Mage replaces makefiles, and is written in Go.
//
// For more detailed information on a task, you can run: mage -h <task> (such as mage -h azure:aksauth).
package main

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/DelineaXPM/dsv-cli/magefiles/constants"
	"github.com/bitfield/script"
	"github.com/dustin/go-humanize"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
	"github.com/sheldonhull/magetools/tooling"

	//mage:import
	_ "github.com/sheldonhull/magetools/gotools"
	//mage:import
	_ "github.com/sheldonhull/magetools/docgen"
	//mage:import
	_ "github.com/DelineaXPM/dsv-cli/magefiles/certs"
)

// relTime returns just a simple relative time humanized, without the "ago" suffix.
func relTime(t time.Time) string {
	return strings.ReplaceAll(humanize.Time(t), " ago", "")
}

// createDirectories creates the local working directories for build artifacts and tooling.
func createDirectories() error {
	magetoolsutils.CheckPtermDebug()
	for _, dir := range []string{constants.ArtifactDirectory} {
		if err := os.MkdirAll(dir, constants.PermissionUserReadWriteExecute); err != nil {
			pterm.Error.Printf("failed to create dir: [%s] with error: %v\n", dir, err)

			return err
		}
		pterm.Success.Printf("✅ [%s] dir created\n", dir)
	}

	return nil
}

// 🧹 Clean up after yourself, artifacts removed, but cache preserved.
func Clean() {
	magetoolsutils.CheckPtermDebug()
	pterm.Info.Println("Cleaning...")
	for _, dir := range []string{constants.ArtifactDirectory} {
		err := sh.Rm(dir)
		if err != nil {
			pterm.Error.Printf("failed to removeall: [%s] with error: %v\n", dir, err)
		}
		pterm.Success.Printf("🧹 [%s] dir removed\n", dir)
	}
	mg.Deps(createDirectories)
}

// 🧹 DeepClean removes both artifacts and cache directory contents.
// Use this when you want to start over including any locally cached certs, files, or other things that normally you'd preserve between test runs.
func DeepClean() {
	magetoolsutils.CheckPtermDebug()
	pterm.Info.Println("🔥 Deep Cleaning...")
	for _, dir := range []string{constants.ArtifactDirectory, constants.CacheDirectory} {
		err := sh.Rm(dir)
		if err != nil {
			pterm.Error.Printf("failed to removeall: [%s] with error: %v\n", dir, err)
		}
		pterm.Success.Printf("🧹 [%s] dir removed\n", dir)
	}
	mg.Deps(createDirectories)
}

func Init() error {
	start := time.Now()
	magetoolsutils.CheckPtermDebug()
	pterm.Success.Println("running Init()...")

	var err error

	if ci.IsCI() {
		pterm.DefaultHeader.Println("CI detected, minimal init being applied")
		pterm.Info.Println("Installing Core CI Dependencies")
		if err = tooling.SilentInstallTools([]string{
			// PRIOR TOOLING - REPLACED BY GOTESTSUM + codecov tooling
			// "github.com/hansboder/gocovmerge@latest",
			// "github.com/jstemmer/go-junit-report/v2@latest",
			// "github.com/axw/gocov/gocov@latest",
			// "github.com/AlekSi/gocov-xml@latest",

			// "github.com/mitchellh/gon/cmd/gon@latest", // macOS binary signing
			// "github.com/miniscruff/changie@latest",    // NOTE: IF run on windows might need this: AS WINDOWS IS NOT WORKING WITH AQUA
			// "github.com/goreleaser/goreleaser@latest", // NOTE: IF run on windows might need this: AS WINDOWS IS NOT WORKING WITH AQUA
			// "github.com/anchore/syft/cmd/syft@latest", // NOTE: IF run on windows might need this: AS WINDOWS IS NOT WORKING WITH AQUA
		}); err != nil {
			return err
		}

		// If goos is windows, then run SilentInstallTools since aqua isn't installing the tools correctly for windows.
		if runtime.GOOS == "windows" {
			if err = tooling.SilentInstallTools([]string{
				"github.com/miniscruff/changie@latest",    // AS WINDOWS IS NOT WORKING WITH AQUA
				"github.com/goreleaser/goreleaser@latest", // AS WINDOWS IS NOT WORKING WITH AQUA
				"github.com/anchore/syft/cmd/syft@latest", // AS WINDOWS IS NOT WORKING WITH AQUA
			}); err != nil {
				return err
			}
		}
		pterm.Success.Printfln("Init() completed (exited early due to ci.IsCI()) [%s]\n", relTime(start))
		return nil
	}

	if runtime.GOOS == "windows" {
		pterm.Warning.Printfln("Trunk is not supported on windows, must run in WSL2, skipping trunk install")
	} else {
		if err = InstallTrunk(); err != nil {
			pterm.Error.Printfln("failed to install trunk (try installing manually from: https://trunk.io/): %v", err)
			return err
		}
	}

	if runtime.GOOS == "darwin" {
		pterm.Info.Println("installing snapcraft for Darwin to allow test build and packaging to run")
		if err := sh.RunWithV(map[string]string{"HOMEBREW_NO_AUTO_UPDATE": "1"}, "brew", "install", "snapcraft"); err != nil {
			pterm.Error.Printfln("failed to install snapcraft on darwin system with homebrew: %v", err)
			return err
		}
		pterm.Success.Println("snapcraft installed")
	}
	// Aqua install is run in devcontainer/codespace automatically.
	// If this environment isn't being used, try to jump start, but if failure, output warning and let the developer choose if they want to go install or not.
	if err := sh.RunV("aqua", "install"); err != nil {
		pterm.Warning.Printfln("aqua install not successful.\n" +
			"This is optional, but will ensure every tool for the project is installed and matching version." +
			"To install see developer docs or go to https://aquaproj.github.io/docs/reference/install")
	}
	pterm.Success.Printfln("Init() completed [%s]\n", relTime(start))
	return nil
}

// InstallTrunk installs trunk.io tooling for linting and formatting.
func InstallTrunk() error {
	_, err := script.Exec("curl https://get.trunk.io -fsSL").Exec("bash -s -- -y").Stdout()
	if err != nil {
		return err
	}

	return nil
}

// TrunkInit ensures the required runtimes are installed.
func TrunkInit() error {
	return sh.RunV("trunk", "install")
}
