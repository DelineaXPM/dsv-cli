// loghelper provides simple helper functions for enabling or disabling more logging with Pterm.
package magetoolsutils

import (
	"os"
	"strconv"

	"github.com/magefile/mage/mg"
	"github.com/pterm/pterm"
)

// checkPtermDebug looks for DEBUG=1 and sets debug level output if this is found to help troubleshoot tasks.
func CheckPtermDebug() { //nolint:cyclop,funlen // cyclop,funlen: i'm sure there's a better way, but for now it works, refactor later // TODO: simplify this env var logic check with helpers
	// var debug bool
	// var err error

	// pterm.Debug.Printfln(
	// 	"\nDEBUG: %v\n"+"SYSTEM_DEBUG: %v\n"+"ACTIONS_STEP_DEBUG: %v\n",
	// 	os.Getenv("DEBUG"),
	// 	os.Getenv("SYSTEM_DEBUG"),
	// 	os.Getenv("ACTIONS_STEP_DEBUG"),
	// )

	// --------------------- GENERAL DEBUG ---------------------
	envDebug, isSet := os.LookupEnv("DEBUG")
	if isSet {
		debug, err := strconv.ParseBool(envDebug)
		if err != nil {
			pterm.Warning.WithShowLineNumber(true).
				WithLineNumberOffset(1).
				Printfln("ParseBool(DEBUG): %v\t debug: %v", err, debug)
		}
		if debug {
			pterm.Debug.Println("strconv.ParseBool(\"DEBUG\") true, enabling debug output and exiting")
			pterm.Debug.Println("DEBUG env var detected, setting tasks to debug level output")
			pterm.EnableDebugMessages()
			return
		}
	}

	// --------------------- AZURE DEVOPS DEBUG ---------------------
	envSystemDebug, isSet := os.LookupEnv("SYSTEM_DEBUG")
	if isSet {
		debug, err := strconv.ParseBool(envSystemDebug) // CI: azure devops uses this for diagnostic level output
		if err != nil {
			pterm.Warning.WithShowLineNumber(true).
				WithLineNumberOffset(1).
				Printfln("ParseBool(SYSTEM_DEBUG): %v\t debug: %v", err, debug)
		}

		if debug {
			pterm.Debug.Println("strconv.ParseBool(\"SYSTEM_DEBUG\") true, enabling debug output and exiting")
			pterm.Debug.Println("SYSTEM_DEBUG env var detected, setting tasks to debug level output")
			pterm.EnableDebugMessages()
			return
		}
	}

	// --------------------- GITHUB ACTIONS DEBUG ---------------------
	envActionsDebug, isSet := os.LookupEnv("ACTIONS_STEP_DEBUG")
	if isSet {
		debug, err := strconv.ParseBool(envActionsDebug) // CI: github uses this for diagnostic level output
		if err != nil {
			pterm.Warning.WithShowLineNumber(true).
				WithLineNumberOffset(1).
				Printfln("ParseBool(ACTIONS_STEP_DEBUG): %v\t debug: %v", err, debug)
		}
		if debug {
			pterm.Debug.Println("strconv.ParseBool(\"ACTIONS_STEP_DEBUG\") true, enabling debug output and exiting")
			pterm.Info.Println("ACTIONS_STEP_DEBUG env var detected, setting tasks to debug level output")
			pterm.EnableDebugMessages()
			return
		}
	}
	if mg.Verbose() {
		pterm.Debug.Printfln("mg.Verbose() true, setting pterm.EnableDebugMessages()")
		pterm.Debug.Println("mg.Verbose() true (-v or MAGEFILE_VERBOSE env var), setting tasks to debug level output")
		pterm.EnableDebugMessages()
		return
	}
}
