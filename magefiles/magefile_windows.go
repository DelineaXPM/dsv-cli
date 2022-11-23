package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
)

// ✏️ Sign will use signtool in Windows to sign the binary.
// Required environment variables will be checked.
func Sign() error {
	var err error
	var errorCount int
	tbl := pterm.TableData{
		[]string{"Status", "Check", "Value", "Notes"},
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

	cliVersion, isSet, tbl := checkEnvVar(
		"CLI_VERSION",
		tbl,
		false,
		"required environment variable",
	)
	if !isSet {
		errorCount++
	}

	cliName, isSet, tbl := checkEnvVar(
		"CLI_NAME",
		tbl,
		false,
		"required environment variable",
	)
	if !isSet {
		errorCount++
	}
	artifactDownloadDirectory, isSet, tbl := checkEnvVar(
		"ARTIFACT_DOWNLOAD_DIRECTORY",
		tbl,
		false,
		"required environment variable",
	)
	if !isSet {
		errorCount++
	}
	certPath, isSet, tbl := checkEnvVar(
		"CERT_PATH",
		tbl,
		false,
		"required cert path",
	)
	if !isSet {
		errorCount++
	}
	certPass, isSet, tbl := checkEnvVar(
		"CERT_PASS",
		tbl,
		true, // secret: don't log
		"required cert pass",
	)
	if !isSet {
		errorCount++
	}
	versionToSign := []string{
		filepath.Join(artifactDownloadDirectory, cliVersion, fmt.Sprintf("%s-win-x64.exe", cliName)),
		filepath.Join(artifactDownloadDirectory, cliVersion, fmt.Sprintf("%s-win-x86.exe", cliName)),
	}
	pterm.Info.Printfln("versionToSign: %+v", versionToSign)
	for _, binary := range versionToSign {
		if _, err := os.Stat(binary); os.IsNotExist(err) {
			pterm.Error.Printfln("%s: does not exist, problem with target path", binary)
			tbl = append(tbl, []string{"❌", binary, "failure", "binary path must be resolved"})
			errorCount++
		} else {
			tbl = append(tbl, []string{"✅", binary, "success", "exists"})
		}
	}
	signTool, err := exec.LookPath("SignTool.exe")
	if err != nil {
		pterm.Error.Println("not able to resolve signtool path, and this is required to run signing, will still attempt to run in case path resolved in system")
		// errorCount++
		tbl = append(tbl, []string{"❌", "SignTool.exe", "failure", "binary path must be resolved"})
	} else {
		tbl = append(tbl, []string{"✅", signTool, "success", "binary found"})
	}

	if errorCount > 0 {
		pterm.Error.Printfln("terminating task since errorCount '%d' exceeds 0", errorCount)
		return fmt.Errorf("terminating task since errorCount '%d' exceeds 0", errorCount)
	}

	for _, binary := range versionToSign {
		if err := sh.RunV(
			signTool,
			"sign",
			"/f", certPath,
			"/p", certPass,
			"/tr",
			"http://timestamp.digicert.com",
			"/td",
			"sha256",
			"/fd",
			"sha256",
			"/sha1",
			"668feb4178afea4d3c4ae833459b09c2bcf6b64e",
			binary,
		); err != nil {
			pterm.Error.Println("unable to sign binary")
			return err
		}
	}

	pterm.Success.Println("Sign()")
	return nil
}
