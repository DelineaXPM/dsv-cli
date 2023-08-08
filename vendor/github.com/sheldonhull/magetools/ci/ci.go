// ci helps identify when a task is running in a ci context and not interactively
// Currently this supports checking:
// 1. Azure DevOps
// 2. GitHub Actions
//
// Any calling packages can just run `isci := ci.IsCI()`
package ci

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// IsCI will set the global variable for IsCI based on lookup of the environment variable.
func IsCI() bool {
	magetoolsutils.CheckPtermDebug()

	_, exists := os.LookupEnv("AGENT_ID")
	if exists {
		pterm.Info.Println("Azure DevOps match based on AGENT_ID. Setting IS_CI = 1")

		return true
	}
	_, exists = os.LookupEnv("GITLAB_CI")
	if exists {
		pterm.Info.Println("Gitlab Runner match based on [GITLAB_CI] env variable. Setting IS_CI = 1")

		return true
	}

	_, exists = os.LookupEnv("NETLIFY")
	if exists {
		pterm.Info.Println("Netlify match based on [NETLIFY] environment. Setting IS_CI = 1")

		return true
	}
	_, exists = os.LookupEnv("GITHUB_ACTIONS")
	if exists {
		pterm.Info.Println("GitHub Actions match based on [GITHUB_ACTIONS] environment. Setting IS_CI = 1")

		return true
	}

	// CI is also set for Netlify, so it's important to run the NETFLIFY check before the CI check.
	// It might be used by others since it's so common so let's leave this check to the very end.
	_, exists = os.LookupEnv("CI")
	if exists {
		pterm.Info.Println("GitHub actions match based on [CI] env variable. Setting IS_CI = 1")

		return true
	}

	return false
}
