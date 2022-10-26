// 🧙 Mage replaces makefiles, and is written in Go.
//
// For more detailed information on a task, you can run: mage -h <task> (such as mage -h azure:aksauth).
package main

import (
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
			"github.com/hansboder/gocovmerge@latest",
			"github.com/jstemmer/go-junit-report/v2@latest",
			"github.com/axw/gocov/gocov@latest",
			"github.com/AlekSi/gocov-xml@latest",
		}); err != nil {
			return err
		}
		return nil
	}
	pterm.Success.Printf("Init() completed [%s]\n", relTime(start))

	return nil
}
