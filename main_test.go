package main

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

var systemTest bool

func init() {
	sysTest := os.Getenv("IS_SYSTEM_TEST")
	if sysTest != "" {
		var err error
		systemTest, err = strconv.ParseBool(sysTest)
		if err != nil {
			systemTest = false
		}
	}
}

// Test started when the generated test binary is started.
// Use the system test flag so this test is not started directly by go test. It is only invoked via the cicd-integration tests
func TestSystem(t *testing.T) {
	if systemTest {
		var args []string
		for a := range os.Args {
			// ignore the -test and coverage arguments, these are for the go test binary and our CLI will throw errors
			// due to unknown flags
			if strings.HasPrefix(os.Args[a], "-test.") || strings.HasSuffix(os.Args[a], ".out") {
				continue
			}
			args = append(args, os.Args[a])
		}
		// use the runCLI method. Calling os.Exit() from a test breaks the test
		_, _ = runCLI(args)
	}
}
