// 🧙 Mage replaces makefiles, and is written in Go.
//
// For more detailed information on a task, you can run: mage -h <task> (such as mage -h azure:aksauth).
package main

import (
	"runtime"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
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
			"github.com/miniscruff/changie@latest",    // AS WINDOWS IS NOT WORKING WITH AQUA
			"github.com/goreleaser/goreleaser@latest", // AS WINDOWS IS NOT WORKING WITH AQUA
			"github.com/anchore/syft/cmd/syft@latest", // AS WINDOWS IS NOT WORKING WITH AQUA
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
	}

	pterm.Success.Printfln("Init() completed [%s]\n", relTime(start))
	return nil
}
