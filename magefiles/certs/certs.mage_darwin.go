package certs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"

	"github.com/DelineaXPM/dsv-cli/magefiles/constants"
)

type Certs mg.Namespace

// [deprecated] Init downloads and installs the required certs for signing binaries from Apple.
// See docs/developer/code-signing.md for more info on Mac local signing.
//
// This isn't currently being used as not signing with Apple based cert approach.
func (Certs) Init() error {
	magetoolsutils.CheckPtermDebug()
	var err error

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	pterm.Debug.Printfln("homeDir: %s", homeDir)
	pterm.Warning.Printfln("This might require interactive approval to allow this to proceed on Darwin based system")

	certsToDownload := []struct {
		URL      string // URL is the file to download.
		FileName string // FileName is the name of the file to save the asset as.
	}{
		{URL: "https://www.apple.com/appleca/AppleIncRootCertificate.cer", FileName: "AppleIncRootCertificate.cer"},
		{URL: "https://developer.apple.com/certificationauthority/AppleWWDRCA.cer", FileName: "AppleWWDRCA.cer"},
	}
	prog, _ := pterm.DefaultProgressbar.
		WithTotal(len(certsToDownload)).
		WithTitle("Downloading Certs").Start()

	defer func() {
		_, _ = prog.Stop()
	}()

	for _, cert := range certsToDownload {
		targetFile := filepath.Join(constants.CacheDirectory, cert.FileName)
		prog.Title = fmt.Sprintf("checking for %s", cert.FileName)
		// Check if the file exists. If not download it, otherwise we'll just install it.
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			// Download the file from the url
			pterm.Debug.Printfln("Downloading %s to %s", cert.URL, targetFile)
			err = sh.RunV("curl", "-o", targetFile, cert.URL)
			if err != nil {
				return err
			}
			pterm.Info.Printfln("downloading cert: %s", targetFile)
		} else {
			pterm.Success.Printfln("cert already downloaded: %s", targetFile)
		}
		prog.Title = "Installing Certs"
		// Install the cert
		err = sh.Run(
			"security",
			"add-trusted-cert",
			"-d",
			"-r",
			"unspecified",
			"-k",
			filepath.Join(homeDir, "Library/Keychains/login.keychain-db"),
			targetFile,
		)
		if err != nil {
			return err
		}
		prog.Increment()
	}

	// If the cer
	return nil
}
