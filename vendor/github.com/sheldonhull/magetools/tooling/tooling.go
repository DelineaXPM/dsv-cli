// Package tooling provides common tooling install setup for go linting, formatting, and other Go tools with nice console output for interactive use.
package tooling

import (
	"bufio"
	"fmt"
	"os/exec"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// InstallTools installs tooling for the project in a local directory to avoid polluting global modules.
func InstallTools(tools []string) error {
	magetoolsutils.CheckPtermDebug()
	pterm.DefaultHeader.Println("InstallTools")
	start := time.Now()

	pterm.DefaultSection.Println("Installing Tools")
	env := map[string]string{}
	args := []string{"install"}

	// spinnerLiveText, _ := pterm.DefaultSpinner.Start("InstallTools")
	defer func() {
		// duration := time.Since(start)
		msg := fmt.Sprintf("tooling installed: %s\n", humanize.Time(start))
		pterm.Success.Println(msg)
		// spinnerLiveText.Success(msg) // Resolve spinner with success message.
	}()

	// as of last time I checked 2021-09, go get/install wasn't noted to be safe to run in parallel, so keeping it simple with just loop
	for idx, tool := range tools {
		msg := fmt.Sprintf("install [%d] %s]", idx, tool)
		// spinner, _ := pterm.DefaultSpinner.
		// 	WithSequence("|", "/", "-", "|", "/", "-", "\\").
		// 	WithRemoveWhenDone(true).
		// 	WithText(msg).
		// 	WithShowTimer(true).Start()

		err := sh.RunWith(env, "go", append(args, tool)...)
		if err != nil {
			pterm.Warning.Printf("Could not install [%s] per [%v]\n", tool, err)
			// spinner.Fail(fmt.Sprintf("Could not install [%s] per [%v]\n", t, err))

			continue
		}
		// spinner.Success(msg)
		pterm.Success.Println(msg)
	}

	return nil
}

// SilentInstallTools reads the stdout and then uses a spinner to show progress.
// This is designed to swallow up a lot of the noise with go install commands.
// Originally found from: https://www.yellowduck.be/posts/reading-command-output-line-by-line/ and modified.
//nolint:funlen // This is ok for now. Can refactor into smaller pieces later if needed.
func SilentInstallTools(toolList []string) error {
	var errorCount int
	magetoolsutils.CheckPtermDebug()
	if ci.IsCI() {
		pterm.DisableStyling()
	}
	pterm.DefaultSection.Println("SilentInstallTools")
	start := time.Now()

	// delay := time.Second * 1 // help prevent jitter
	spin, _ := pterm.DefaultSpinner. // WithDelay((delay)).WithRemoveWhenDone(true).
						WithShowTimer(true).
						WithText("go install tools").
						WithSequence("|", "/", "-", "|", "/", "-", "\\").
						Start()
		// WithSequence("|", "/", "-", "|", "/", "-", "\\").

	// spinnerLiveText, _ := pterm.DefaultSpinner.Start("InstallTools")

	pterm.Info.Printf("items to iterate through: %d", len(toolList))
	for _, item := range toolList {
		cmd := exec.Command("go", "install", item)

		status := "go install " + item
		// Get a pipe to read from standard out
		r, _ := cmd.StdoutPipe()

		// Use the same pipe for standard error
		cmd.Stderr = cmd.Stdout

		// Make a new channel which will be used to ensure we get all output
		done := make(chan struct{})

		// Create a scanner which scans r in a line-by-line fashion
		scanner := bufio.NewScanner(r)

		// Use the scanner to scan the output line by line and log it
		// It's running in a goroutine so that it doesn't block
		go func(status string, spin *pterm.SpinnerPrinter) {
			// Read line by line and process it
			spin.UpdateText(status)
			for scanner.Scan() {
				line := scanner.Text()
				spin.UpdateText(line)
			}
			// We're all done, unblock the channel
			done <- struct{}{}
		}(status, spin)

		// Start the command and check for errors
		err := cmd.Start()
		if err != nil {
			pterm.Error.Printfln("unable to install: %q %v", item, err)
			errorCount++
			spin.Fail(err)
			continue
			// _ = spin.Stop()  // NOTE: continue installing other tools, don't stop everything, just fail this and count it
			// return err
		}

		// Wait for all output to be processed
		<-done

		// Wait for the command to finish
		err = cmd.Wait()
		if err != nil {
			spin.Fail(err)
			errorCount++
			pterm.Error.Printfln("unable to install: %q %v", item, err)
			// _ = spin.Stop() // NOTE: continue installing other tools, don't stop everything, just fail this and count it
			// return err
			continue
		}
		spin.Success(item)
	}

	if errorCount > 0 {
		pterm.Error.Printfln("SilentInstallTools: total errors: [%d] %s\n", errorCount, humanize.Time(start))
		return fmt.Errorf("SilentInstallTools: total errors: [%d]", errorCount)
	}
	msg := fmt.Sprintf("SilentInstallTools: %s\n", humanize.Time(start))
	pterm.Success.Println(msg)
	return nil
}

// SilentInstallTools reads the stdout and then uses a spinner to show progress.
// binary: name of tool to run, such as go, gofumpt, etc.
//
// cmdargs: slice of string arguments to pass into command
//
// list: optional list of strings that the command will iterate against.
//
// - If no arguments, will range 1x over blank string
//
// - If list is provided then loop will provide each invocation against it.
//
// Example: SpinnerStdOut("go",[]string{"mod","tidy"},[]string{""})
//
// Example: SpinnerStdOut("go",[]string{"install"},[]string{	"golang.org/x/tools/cmd/goimports@master","github.com/sqs/goreturns@master"})
// This is designed to swallow up a lot of the noise with go install commands.
// Originally found from: https://www.yellowduck.be/posts/reading-command-output-line-by-line/ and modified.
//nolint:funlen // Bypassing. Will need to evaluate later if I want to break this apart. For now it's not important
func SpinnerStdOut(
	binary string,
	cmdargs, list []string,
) error {
	magetoolsutils.CheckPtermDebug()
	pterm.DefaultHeader.Println(fmt.Sprintf("%s %v", binary, cmdargs))
	// delay := time.Second * 1 // help prevent jitter
	start := time.Now()
	spin, _ := pterm.DefaultSpinner. // WithDelay((delay)).WithRemoveWhenDone(true).
						WithShowTimer(true).
						WithText(fmt.Sprintf("%s %v", binary, cmdargs)).
						Start()
		// WithSequence("|", "/", "-", "|", "/", "-", "\\").

	defer func() {
		msg := fmt.Sprintf("%s: %q %s\n", binary, cmdargs, humanize.Time(start))
		spin.Success(msg) // Resolve spinner with success message.
		// pterm.Success.Println(msg)
	}()
	// NOTE: if empty set string to "empty" so iteration 1 time will work.
	// Not sure if alternative, will revisit in future.
	if len(list) == 0 {
		pterm.Debug.Println("set list to \"empty\" to force iteration 1x")
		list = []string{"empty"}
	}
	pterm.Info.Printf("items to iterate through: %d", len(list))
	for _, item := range list[0:] {
		pterm.Debug.Printf("item: %s\n", item)
		thisargs := []string{}
		thisargs = append(thisargs, cmdargs...)
		var status string
		// takes the cmdargs and appends the item to iterate on to the end of the slice
		// this would then mean go install x becomes go install repo@latest
		//
		// If blank list is provided then this would be go mod tidy <nil> and no invalid string arg would be passed
		status = item // default to item, but override value with cmd if no args are provided
		if item == "empty" {
			item = ""
			pterm.Debug.Println(
				"item matched \"empty\" so inside range loop I'm setting now to empty",
			)
			status = fmt.Sprintf("%s %q", binary, thisargs)
		}
		if item != "" {
			thisargs = append(thisargs, item)
			pterm.Debug.Printf("item: %q not nil, so adding to cmd\n", item)
			status = fmt.Sprintf("%s %q", binary, thisargs)
		}
		pterm.Debug.Printf("exec.Command(%s, %q)\n", binary, thisargs)
		time.Sleep(time.Second)
		cmd := exec.Command(binary, thisargs...)

		// Get a pipe to read from standard out
		readCloser, _ := cmd.StdoutPipe()

		// Use the same pipe for standard error
		cmd.Stderr = cmd.Stdout

		// Make a new channel which will be used to ensure we get all output
		done := make(chan struct{})

		// Create a scanner which scans r in a line-by-line fashion
		scanner := bufio.NewScanner(readCloser)

		// Use the scanner to scan the output line by line and log it
		// It's running in a goroutine so that it doesn't block
		go func(status string, spin *pterm.SpinnerPrinter) {
			// Read line by line and process it
			spin.UpdateText(status)
			for scanner.Scan() {
				line := scanner.Text()
				spin.UpdateText(line)
			}
			// We're all done, unblock the channel
			done <- struct{}{}
		}(status, spin)

		// Start the command and check for errors
		err := cmd.Start()
		if err != nil {
			spin.Fail(err)
			_ = spin.Stop()
			return err
		}

		// Wait for all output to be processed
		<-done

		// Wait for the command to finish
		err = cmd.Wait()
		if err != nil {
			spin.Fail(err)
			_ = spin.Stop()
			return err
		}
		spin.Success(item)
	}

	return nil
}
