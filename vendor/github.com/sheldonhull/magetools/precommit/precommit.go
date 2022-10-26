// Package precommit provides pre-commit style setup and hook registration.
//
// Install pre-commit in your project with asdf.
// Just run `asdf plugin-add pre-commit && asdf install pre-commit latest && asdf local pre-commit latest`
// At this time the package focuses on pre-commit framework, since it's well established.
//
// Lefthook is also great, but requires a lot more effort and custom work, as well as implements less cross-platform normalization, so pre-commit is the tool of choice here.
//
// Links:
//
// - https://pre-commit.com/
//
//
//

package precommit

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"regexp"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// Precommit contains tasks for Pre-commit for pre-commit/pre-push automation.
type Precommit mg.Namespace

// ⚙️ Init configures precommit hooks.
func (Precommit) Init() error {
	magetoolsutils.CheckPtermDebug()

	pterm.DefaultHeader.Println("⚡ Registering pre-commit")

	var err error
	var python3 string

	_, err = os.Stat(".pre-commit-config.yaml")
	if err != nil {
		_, err := ProvideDefaultConfig()
		if err != nil {
			pterm.Warning.Printfln("unable to provide a default pre-commit configuration: %v", err)
		}
	}
	_, err = exec.LookPath("python3")

	// If python3 isn't found, let's try to resolve that with some magic.
	if os.IsNotExist(err) {
		pterm.Error.Println("unable to find python3, this might not be installed, or available in path")
		if err != nil {
			pterm.Error.Println("unable to resolve python3 still. you'll have to do this on your own")
			return err
		}
	}

	r := regexp.MustCompile(`executable file not found in $PATH`)
	if err := sh.RunV(python3, "-m", "pip", "install", "pre-commit", "--user"); err != nil {
		if r.MatchString(err.Error()) {
			pterm.Error.Println(
				"pip not found, so something didn't get setup right with python3. you'll have to fix this yourself unfortunately. 💔",
			)
			return err
		}
		if err := sh.RunV("pre-commit", "install"); err != nil {
			return err
		}
		pterm.Success.Println("✅ registered pre-commit")
		return nil
	}
	return sh.RunV("pre-commit", "install")
}

// 🧪 Commit runs pre-commit checks using pre-commit.
//
// If using pre-commit framework you can see the available stages on the [docs](https://pre-commit.com/#confining-hooks-to-run-at-certain-stages).
//
// > "The stages property is an array and can contain any of commit, merge-commit, push, prepare-commit-msg, commit-msg, post-checkout, post-commit, post-merge, post-rewrite, and manual.".
func (Precommit) Commit() error {
	magetoolsutils.CheckPtermDebug()
	return sh.RunV("pre-commit", "run", "--hook-stage", "pre-commit")
}

// 🧪 Push runs pre-push checks using pre-commit.
func (Precommit) Prepush() error {
	magetoolsutils.CheckPtermDebug()
	return sh.RunV("pre-commit", "run", "--hook-stage", "push")
}

// ProvideDefaultConfig will render a default pre-commit yaml file if doesn't exist to bootstrap the project.
func ProvideDefaultConfig() (string, error) {
	templatefile := "precommit.data.tmpl"
	outfile := ".pre-commit-config.yaml"

	f, err := os.Create(outfile)
	if err != nil {
		return "", fmt.Errorf("os.Create(outfile): %w", err)
	}
	t, err := template.ParseFiles(templatefile)
	if err != nil {
		return "", fmt.Errorf("template.ParseFiles(templatefile): %w", err)
	}
	err = t.Execute(f, "precommit")
	if err != nil {
		f.Close()
		pterm.Error.Printfln("unable to generate %s: %v", outfile, err)
		return "", fmt.Errorf("t.Execute(f, \"precommit\"%w", err)
	}

	f.Close()
	return outfile, nil
}

// ✖ Uninstall removes the pre-commit hooks.
func (Precommit) Uninstall() error {
	magetoolsutils.CheckPtermDebug()
	return sh.RunV("pre-commit", "uninstall")
}
