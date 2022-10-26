// Package req provides methods to resolve Go oriented tooling paths, and if not found attempts to install on demand.
//
// This simplifies other packages so they don't need to worry about installing tools each time.
// Instead the packages get installed on demand when called.
//
// Example:
//
// Let's say you run mage secrets:check but don't have gitleaks installed.
//
// The package tasks will run the check, but if the binary for gitleaks isn't found, then it would attempt to run the `go install github.com/zricethezav/gitleaks/v8` command, resolve the path, and provide this pack to the caller.
//
// Overtime, I've started migrating more to this approach as it means you have far less concerns for tools like this to run any install/init style setup, and instead just let it self-setup as needed.
package req

import (
	"go/build"
	"os/exec"
	"strings"

	"github.com/bitfield/script"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/tooling"
	"github.com/ztrue/tracerr"
)

// GetGoPath returns the GOPATH value.
func GetGoPath() (gopath string) {
	p := script.Exec("go env GOPATH")
	s, _ := p.String()
	gopath = strings.TrimSpace(s)
	pterm.Debug.Printfln("GOPATH pulled from `go env GOPATH`: %s", gopath)
	if gopath == "" {
		gopath = build.Default.GOPATH
		pterm.Debug.Printfln("GOPATH not found from `go env GOPATH` so using build.Default.GOPATH: %s", gopath)
	}
	return gopath
}

// ResolveBinaryByInstall tries to qualify the tool, looking in all `PATH` locations.
// If it can't resolve the binary, it will attempt to use a source based install via `go install`, which might be slower, but eliminate failure if not using a pre-installation method such as Aqua, asdf, brew, or other tooling methods.
//
// This can help with running in CI and not having to have a lot of setup code.
func ResolveBinaryByInstall(app, goInstallCmd string) (qualifiedBinary string, err error) {
	qualifiedBinary, err = QualifyGoBinary(app)

	if err != nil {
		pterm.Info.Printfln("Couldn't find %s, so will attempt to install it from source", app)
		err := tooling.SilentInstallTools([]string{goInstallCmd})
		if err != nil {
			return "", tracerr.Wrap(err)
		}

		qualifiedBinary, err = QualifyGoBinary(app)
		if err != nil {
			pterm.Error.WithShowLineNumber(true).
				WithLineNumberOffset(1).
				Printfln("second try to QualifyGoBinary failed: %v", err)
			return "", tracerr.Wrap(err)
		}
	}
	return qualifiedBinary, nil
}

// addGoPkgBinToPath ensures the go/bin directory is available in path for cli tooling.
// This isn't used right now as I prefer to use fully qualified tool paths which don't care about env var issues.
// func addGoPkgBinToPath() error {
// 	gopath := GetGoPath()
// 	goPkgBinPath := filepath.Join(gopath, "bin")
// 	if !strings.Contains(os.Getenv("PATH"), goPkgBinPath) {
// 		pterm.Debug.Printf("Adding %q to PATH\n", goPkgBinPath)
// 		updatedPath := strings.Join([]string{goPkgBinPath, os.Getenv("PATH")}, string(os.PathListSeparator))
// 		if err := os.Setenv("PATH", updatedPath); err != nil {
// 			pterm.Error.WithShowLineNumber(true).WithLineNumberOffset(1).Printfln("Error setting PATH: %v\n", err)
// 			return tracerr.Wrap(err)
// 		}
// 		pterm.Info.Printf("Updated PATH: %q\n", updatedPath)
// 	}
// 	pterm.Debug.Printf("bypassed PATH update as already contained %q\n", goPkgBinPath)
// 	return nil
// }

// QualifyGoBinary provides a fully qualified path for an installed Go binary to avoid path issues.
// This uses exec.LookPath to allow resolution of the binary from any provided `PATH` variables, allowing better alternative tooling installation before installing from source.
func QualifyGoBinary(binary string) (string, error) {

	pterm.Debug.Printfln("searching for binary in all provided PATH locations: %q", binary)
	qualifiedPath, err := exec.LookPath(binary)
	if err != nil {
		pterm.Warning.WithShowLineNumber(true).
			WithLineNumberOffset(1).
			Printfln("%q not found in any provided locations", binary)
		return "", tracerr.Wrap(err)
	}
	pterm.Debug.Printfln("%q full path: %q", binary, qualifiedPath)
	return qualifiedPath, nil
}
