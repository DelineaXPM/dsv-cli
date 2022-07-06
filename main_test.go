package main

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// Test started when the generated test binary is started.
// Use the system test flag so this test is not started directly by go test.
func TestSystem(t *testing.T) {
	sysTest := os.Getenv("IS_SYSTEM_TEST")
	if sysTest == "" {
		return
	}

	systemTest, err := strconv.ParseBool(sysTest)
	if err != nil || !systemTest {
		return
	}

	var args []string
	for _, arg := range os.Args {
		// Ignore the -test and coverage arguments, these are for the go test binary
		// and our CLI will throw errors due to unknown flags.
		if strings.HasPrefix(arg, "-test.") || strings.HasSuffix(arg, ".out") {
			continue
		}

		args = append(args, arg)
	}

	// use the runCLI method. Calling os.Exit() from a test breaks the test
	_, _ = runCLI(args)
}
