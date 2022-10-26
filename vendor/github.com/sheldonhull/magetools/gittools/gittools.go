// Package gittools provides automatic setup of some useful git tooling like Bit and Git Town
package gittools

import (
	"github.com/magefile/mage/mg"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
	"github.com/sheldonhull/magetools/tooling"
)

type Gittools mg.Namespace

// golang tools to ensure are locally vendored.
var toolList = []string{ //nolint:gochecknoglobals // ok to be global for tooling setup
	"github.com/git-town/git-town@latest",
	"github.com/chriswalz/bit@latest",
}

// ⚙️  Init runs all required steps to use this package.
func (Gittools) Init() error {
	magetoolsutils.CheckPtermDebug()
	pterm.DefaultHeader.Println("Gittools Init()")

	if err := tooling.SilentInstallTools(toolList); err != nil {
		return err
	}
	pterm.Info.Println("🔧 Installed tools. To setup aliases automatically run these statements in your terminal:")
	pterm.Info.Println("bit aliases:     \tbit complete")
	pterm.Info.Println("git-town aliases:\tgit town alias true")

	return nil
}
