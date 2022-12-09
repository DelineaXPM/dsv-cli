// This contains customized goreleaser tasks that take into account the GOOS and combine this with my standard approach of using changelog to drive the new semver release.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DelineaXPM/dsv-cli/magefiles/constants"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Build contains mage tasks specific to building without a release.
type (
	Build mg.Namespace
	// Release contains mage tasks specific to the release process, including upload of assets to s3, github, etc.
	Release mg.Namespace
)

// checkEnv is the struct to pass into the checkEnvVar function to check and validate the environment variables.
// This builds a nice table summary when used to help summarize all the failed checks rather than doing this piecemeal.
type checkEnv struct {
	Name       string
	IsSecret   bool
	IsRequired bool
	Tbl        pterm.TableData
	Notes      string
}

// checkEnvVar performs a check on environment variable and helps build a report summary of the failing conditions, missing variables, and bypasses logging if it's a secret.
// Yes this could be replaced by the `env` package but I had this in place and the output is nice for debugging so I left it. - Sheldon 😀
//
//nolint:unparam,appendAssign // ignoring as i'll want to use the values in the future, ok to leave for now. appendAssign is ok as well, though it could be much cleaner, it works for now.
func checkEnvVar(ckv checkEnv) (string, pterm.TableData, error) {
	// loggedValue is used to make sure any secret isn't put into the table output.
	var value, loggedValue string
	var isSet bool
	tbl := ckv.Tbl
	value, isSet = os.LookupEnv(ckv.Name)

	if ckv.IsSecret {
		loggedValue = "***** secret set, but not logged *****"
	} else {
		loggedValue = value
	}

	// Required but not set is an error condition to report back to the user.
	if !isSet && ckv.IsRequired {
		// trunk-ignore(golangci-lint/gocritic)
		tbl = append(ckv.Tbl, []string{"❌", ckv.Name, loggedValue, ckv.Notes})
		return "", tbl, fmt.Errorf("%s is required and not set", ckv.Name)
	}
	// Required but not a terminating error, then just put as information different from success, and no error.
	if !isSet && !ckv.IsRequired {
		// trunk-ignore(golangci-lint/gocritic)
		tbl = append(ckv.Tbl, []string{"👉", ckv.Name, loggedValue, ckv.Notes}) //nolint:appendAssign // might refactor in future, for now it's ok - sheldon
		return value, tbl, nil
	}

	if isSet {
		// trunk-ignore(golangci-lint/gocritic)
		tbl = append(ckv.Tbl, []string{"✅", ckv.Name, loggedValue, ckv.Notes}) //nolint:appendAssign // might refactor in future, for now it's ok - sheldon
		return value, tbl, nil
	}
	return "", tbl, fmt.Errorf("unknown error (no conditions were hit so it's a PEKAB issue 😁) with evaluation of: %s", ckv.Name)
}

// func checkEnvVar(envVar string, required bool) (string, error) { //nolint:unused // leaving this as will use in future releases
// 	envVarValue := os.Getenv(envVar)
// 	if envVarValue == "" && required {
// 		pterm.Error.Printfln(
// 			"%s is required and unable to proceed without this being provided. terminating task.",
// 			envVar,
// 		)
// 		return "", fmt.Errorf("%s is required", envVar)
// 	}
// 	if envVarValue == "" {
// 		pterm.Debug.Printfln(
// 			"checkEnvVar() found no value for: %q, however this is marked as optional, so not exiting task",
// 			envVar,
// 		)
// 	}
// 	pterm.Debug.Printfln("checkEnvVar() found value: %q=%q", envVar, envVarValue)
// 	return envVarValue, nil
// }

// 🔨 Single builds the project for the current platform.
func (Build) Single() error {
	magetoolsutils.CheckPtermDebug()
	releaserArgs := []string{
		"build",
		"--rm-dist",
		"--snapshot",
		"--single-target",
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)

	return sh.RunV("goreleaser", releaserArgs...) // "--skip-announce",.
}

// 🔨 All builds all the binaries defined in the project, for all platforms. This includes Docker image generation but skips publish.
// If there is no additional platforms configured in the task, then basically this will just be the same as `mage build`.
func (Build) All() error {
	magetoolsutils.CheckPtermDebug()
	releaserArgs := []string{
		"release",
		"--snapshot",
		"--rm-dist",
		"--skip-publish",
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)
	return sh.RunV("goreleaser", releaserArgs...)
	// To pass in explicit version mapping, you can do this. I'm not using at this time.
	// Return sh.RunWithV(map[string]string{
	// 	"GORELEASER_CURRENT_TAG": "latest",
	// }, binary, releaserArgs...)
}

// 🔨 All generates a release with goreleaser. This does the whole shebang, including build, publish, and notify.
func (Release) All() error {
	magetoolsutils.CheckPtermDebug()
	// opting to always remove after running release to avoid possible non-snapshot artifact persisting.
	defer func() {
		err := sh.Rm(constants.TargetCLIVersionArtifact)
		if err != nil {
			pterm.Error.Printfln("error removing %s: %v", constants.TargetCLIVersionArtifact, err)
		}
	}()
	// TODO: this will be checked once we publish cli to github
	// if _, err := checkEnvVar("DOCKER_ORG", true); err != nil {
	// 	return err
	// }
	// Run any dependent tasks first.
	mg.SerialDeps(Release{}.GenerateCLIVersionFile)

	releaseVersion, err := sh.Output("changie", "latest")
	if err != nil {
		pterm.Error.Printfln("changie pulling latest release note version failure: %v", err)
		return err
	}
	cleanVersion := strings.TrimSpace(releaseVersion)
	cleanpath := filepath.Join(".changes", cleanVersion+".md")
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		cleanpath = filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".changes", cleanVersion+".md")
	}
	// NOTE: Merging all of this into a single goreleaser, not per-platform anymore.

	releaserArgs := []string{
		"release",
		"--rm-dist",
		"--skip-validate",
		fmt.Sprintf("--release-notes=%s", cleanpath),
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)

	if err := sh.RunWithV(map[string]string{
		"GORELEASER_CURRENT_TAG": cleanVersion,
	},
		"goreleaser",
		releaserArgs...,
	); err != nil {
		return err
	}
	pterm.Println("(Release).All() completed successfully")
	return nil
}

// getGoreleaserConfig returns the path to the goreleaser config file based on the current OS.
func getGoreleaserConfig() (string, error) {
	magetoolsutils.CheckPtermDebug()

	var configfile string
	switch runtime.GOOS {
	case "darwin":
		configfile = ".goreleaser.darwin.yaml"
	case "linux":
		configfile = ".goreleaser.linux.yaml"
	case "windows":
		configfile = ".goreleaser.windows.yaml"
	default:
		pterm.Error.Printfln("Unsupported OS: %s", runtime.GOOS)
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	pterm.Info.Printfln("using config file: %s", configfile)
	return configfile, nil
}

// GenerateCLIVersionFile generates a json object with an array of the containing a list of all the artifact versions and their links based on our standard download url.
func (Release) GenerateCLIVersionFile() error {
	magetoolsutils.CheckPtermDebug()

	releaseVersion, _, err := getVersion()
	if err != nil {
		return err
	}
	// Links is the url for all the assets published.
	//nolint:tagliatelle // this is specifically what the CLI requires.
	type Links struct {
		DarwinAmd64  string `json:"darwin/amd64"`
		DarwinArm64  string `json:"darwin/arm64"`
		LinuxAmd64   string `json:"linux/amd64"`
		Linux386     string `json:"linux/386"`
		WindowsAmd64 string `json:"windows/amd64"`
		Windows386   string `json:"windows/386"`
	}

	// cliVersions is the struct that will be turned into a json file.
	type cliVersions struct {
		Latest string `json:"latest"`
		Links  Links  `json:"links"`
	}

	newJSON := cliVersions{
		Latest: releaseVersion,
		Links: Links{
			DarwinAmd64:  fmt.Sprintf(constants.DownloadURLFString, releaseVersion, "darwin-x64"),
			DarwinArm64:  fmt.Sprintf(constants.DownloadURLFString, releaseVersion, "darwin-arm64"),
			LinuxAmd64:   fmt.Sprintf(constants.DownloadURLFString, releaseVersion, "linux-x64"),
			Linux386:     fmt.Sprintf(constants.DownloadURLFString, releaseVersion, "linux-x86"),
			WindowsAmd64: fmt.Sprintf(constants.DownloadURLFString, releaseVersion, "windows-x64"),
			Windows386:   fmt.Sprintf(constants.DownloadURLFString, releaseVersion, "windows-x86"),
		},
	}

	if err := os.MkdirAll(constants.ArtifactDirectory, constants.PermissionUserReadWriteExecute); err != nil {
		return err
	}
	// Write the json file.
	jf, err := os.Create(constants.TargetCLIVersionArtifact)
	if err != nil {
		pterm.Error.Printfln("error creating json file: %v", err)
		return err
	}

	b, err := json.MarshalIndent(newJSON, "", "  ")
	if err != nil {
		pterm.Error.Printfln("error marshaling json: %v", err)
		return err
	}
	if _, err := jf.Write(b); err != nil {
		pterm.Error.Printfln("error writing json file: %v", err)
		return err
	}
	pterm.Success.Printfln("json file written: %s", jf.Name())

	return nil
}

// getVersion returns the version and path for the changefile to use for the semver and release notes.
func getVersion() (releaseVersion, cleanPath string, err error) {
	releaseVersion, err = sh.Output("changie", "latest")
	if err != nil {
		pterm.Error.Printfln("changie pulling latest release note version failure: %v", err)
		return "", "", err
	}
	cleanVersion := strings.TrimSpace(releaseVersion)
	cleanPath = filepath.Join(".changes", cleanVersion+".md")
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		cleanPath = filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".changes", cleanVersion+".md")
	}
	return cleanVersion, cleanPath, nil
}

// UploadCLIVersion uploads the cli-versions.json file to the secrets s3 bucket.
func (Release) UploadCLIVersion() error {
	// BucketInQuestion contains S3Client, an Amazon S3 service client that is used to perform bucket
	// and object actions.
	//
	// Example from aws https://github.com/awsdocs/aws-doc-sdk-examples/blob/f45333bde292926451ba626e17be1c6a49c037f6/gov2/s3/actions/bucket_basics.go#LL103-L120
	mg.Deps(Release{}.GenerateCLIVersionFile)
	type BucketInQuestion struct {
		S3Client *s3.Client
	}

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(constants.AWSDefaultS3Region))
	if err != nil {
		return err
	}
	bucket := BucketInQuestion{
		S3Client: s3.NewFromConfig(cfg),
	}
	file, err := os.Open(constants.TargetCLIVersionArtifact)
	if err != nil {
		pterm.Error.Printfln("Couldn't open file %v to upload. Here's why: %v", constants.TargetCLIVersionArtifact, err)
		return err
	} else {
		defer file.Close()
		_, err := bucket.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(constants.S3CLIVersionPath),
			Body:   file,
		})
		if err != nil {
			pterm.Error.Printfln("Couldn't upload file %v to %v:%v. Here's why: %v",
				constants.TargetCLIVersionArtifact,
				os.Getenv("S3_BUCKET"), constants.S3CLIVersionPath,
				err,
			)
			return err
		}
	}
	pterm.Success.Printfln("(Release) successfully uploaded file %v to %v:%v",
		constants.TargetCLIVersionArtifact,
		os.Getenv("S3_BUCKET"), constants.S3CLIVersionPath,
	)
	return nil
}
