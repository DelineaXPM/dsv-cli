// Package docs runs documentation tools to generate docs for the project.
// This includes gomarkdoc project documentation.

package docgen

import (
	"fmt"

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

		return err
	}
	pterm.Success.Println("Docs")

	return nil
}
