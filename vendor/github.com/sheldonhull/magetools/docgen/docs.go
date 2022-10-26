// Package docs runs documentation tools to generate docs for the project.
// This includes gomarkdoc project documentation.

package docgen

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
	"github.com/sheldonhull/magetools/tooling"
)

// Docs contains all the Doc related tasks.
type Docs mg.Namespace

// docsDir is the directory where the licenses are stored.
const docsDir = "docs/godocs"

// FullPermissions sets permissions based on same model as gomarkdoc.
const FullPermissions = 0o777

// toolList is a list of tooling to install for the project commands.
var toolList = []string{ //nolint:gochecknoglobals // ok to be global for tooling setup
	"github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest",
}

// ⚙️ Init initializes the tooling for Docs.
func (Docs) Init() error {
	magetoolsutils.CheckPtermDebug()
	if err := tooling.InstallTools(toolList); err != nil {
		return err
	}
	if err := os.MkdirAll(docsDir, FullPermissions); err != nil {
		return fmt.Errorf("unable to create the target directory: %w", err)
	}
	pterm.Success.Printfln("()Init mkdir: %s", docsDir)
	return nil
}

// 📘 Docs generate md docs. Required: Format (azure-devops, github, gitlab).
func (Docs) Generate(format string) error {
	magetoolsutils.CheckPtermDebug()
	pterm.Info.Println("generate Go docs for the project.")
	cmdArgs := []string{
		// "--include-unexported",
		"--format",
		format,
		"--output",
		fmt.Sprintf("%s/{{ .Dir }}.md", docsDir),
		"./...",
	}
	if err := sh.Run("gomarkdoc", cmdArgs...); err != nil {
		pterm.Error.Println(err)
		pterm.Warning.Println(
			"if failure is due to permissions, try running with sudo.\nThis seems to resolve until upstream fix can be placed on gomarkdoc tool.",
		)
		pterm.Warning.Printfln("\tsudo gomarkdoc --format %s --output 'docs/godocs/{{ .Dir }}.md' ./...", format)
		pterm.Warning.Println(
			"\nYou may need to set permissions if they got created by tool incorrectly to open in ide: sudo chmod -R 0777 docs/",
		)
		return err
	}
	pterm.Success.Println("Docs")

	return nil
}
