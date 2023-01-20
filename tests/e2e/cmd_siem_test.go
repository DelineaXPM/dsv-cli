//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"runtime"
	"testing"
)

func TestSIEM_CRUD(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Sorry, interactive End-to-End tests cannot be executed on Windows.")
	}
	e := newEnv()

	var (
		siemName          = makeSIEMName()
		siemHost          = "127.0.0.1"
		siemPort          = "3131"
		siemAuth          = "123"
		siemPool          = ""
		siemAuthType      = "token"
		siemLoggingFormat = "rfc5424"
		siemProtocol      = "udp"
		siemSendToEngine  = "false"
		siemType          = "syslog"
	)

	cmd := []string{
		"siem", "create",
		"--auth-type=password",
		fmt.Sprintf("--auth-username=%s", e.username),
		fmt.Sprintf("--auth-password=%s", e.password),
		fmt.Sprintf("--tenant=%s", e.tenant),
		fmt.Sprintf("--domain=%s", e.domain),
	}

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Name of SIEM endpoint")
		c.SendLine(siemName)
		c.ExpectString("Select SIEM type")
		c.SendKeyEnter()
		c.ExpectString("Select protocol for syslog SIEM type")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()
		c.ExpectString("Host")
		c.SendLine(siemHost)
		c.ExpectString("Port")
		c.SendLine(siemPort)
		c.ExpectString("Select authentication method")
		c.SendKeyEnter()
		c.ExpectString("Authentication")
		c.SendLine(siemAuth)
		c.ExpectString("Select logging format")
		c.SendKeyEnter()
		c.ExpectString("Route through DSV engine")
		c.SendKeyEnter()

		c.ExpectEOF()
	})

	output := runWithAuth(t, e, fmt.Sprintf("siem read %s", siemName))
	requireContains(t, output, siemName)
	requireContains(t, output, siemAuthType)
	requireContains(t, output, siemHost)
	requireContains(t, output, siemPort)
	requireContains(t, output, siemAuth)
	requireContains(t, output, siemLoggingFormat)
	requireContains(t, output, siemProtocol)
	requireContains(t, output, siemSendToEngine)
	requireContains(t, output, siemType)
	requireContains(t, output, siemPool)

	cmd = []string{
		"siem", "update", siemName,
		"--auth-type=password",
		fmt.Sprintf("--auth-username=%s", e.username),
		fmt.Sprintf("--auth-password=%s", e.password),
		fmt.Sprintf("--tenant=%s", e.tenant),
		fmt.Sprintf("--domain=%s", e.domain),
	}

	// Update port in SIEM config.
	siemPort = "3030"

	runFlow(t, cmd, func(c console) {
		c.ExpectString("Select SIEM type")
		c.SendKeyEnter()
		c.ExpectString("Select protocol for syslog SIEM type")
		c.SendKeyArrowDown()
		c.SendKeyArrowDown()
		c.SendKeyEnter()
		c.ExpectString("Host")
		c.SendLine(siemHost)
		c.ExpectString("Port")
		c.SendLine(siemPort)
		c.ExpectString("Select authentication method")
		c.SendKeyEnter()
		c.ExpectString("Authentication")
		c.SendLine(siemAuth)
		c.ExpectString("Select logging format")
		c.SendKeyEnter()
		c.ExpectString("Route through DSV engine")
		c.SendKeyEnter()

		c.ExpectEOF()
	})

	output = runWithAuth(t, e, fmt.Sprintf("siem read %s", siemName))
	requireContains(t, output, siemName)
	requireContains(t, output, siemAuthType)
	requireContains(t, output, siemHost)
	requireContains(t, output, siemPort)
	requireContains(t, output, siemAuth)
	requireContains(t, output, siemLoggingFormat)
	requireContains(t, output, siemProtocol)
	requireContains(t, output, siemSendToEngine)
	requireContains(t, output, siemType)
	requireContains(t, output, siemPool)

	output = runWithAuth(t, e, fmt.Sprintf("siem delete %s", siemName))
	if output != "" {
		t.Fatalf("Unexpected output on delete: \n%s\n", output)
	}
}
