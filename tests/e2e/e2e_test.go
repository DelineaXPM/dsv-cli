//go:build endtoend
// +build endtoend

package e2e

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	expect "github.com/Netflix/go-expect"
	pseudotty "github.com/creack/pty"
	"github.com/hinshun/vt10x"
)

const info = `
 ______ ___  ______ 
|  ____|__ \|  ____|        Go version:  %s
| |__     ) | |__           OS/Arch:     %s/%s
|  __|   / /|  __|          Working dir: %s
| |____ / /_| |____         Time:        %s
|______|____|______|
`

const help = `
===================================================================================================
DevOps Secrets Vault CLI End-to-End testing requires a tenant and an access to it.

A template with all required data is defined in the <project-root>/tests/e2e/.env.example file.

How to run E2E tests:
    - from project root directory create a copy of the example file:
        $ cp ./tests/e2e/.env.example ./tests/e2e/.env
    - fill in all variables in the created .env file
    - execute it:
        $ source ./tests/e2e/.env
    - run E2E tests:
        $ make e2e-test
===================================================================================================

%s

`

const (
	domainEnvName      = "DSV_CLI_E2E_DOMAIN"
	tenantEnvName      = "DSV_CLI_E2E_TENANT"
	usernameEnvName    = "DSV_CLI_E2E_USERNAME"
	passwordEnvName    = "DSV_CLI_E2E_PASSWORD"
	certificateEnvName = "DSV_CLI_E2E_CERTIFICATE"
	privateKeyEnvName  = "DSV_CLI_E2E_PRIVKEY"
)

// Vars managed by TestMain function.
var (
	binPath    = ""
	tmpDirPath = ""
	covDirPath = ""
)

func TestMain(m *testing.M) {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] Unable to determine working directory: %v.\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, info,
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
		workDir, time.Now().Format(time.ANSIC),
	)

	if os.Getenv(domainEnvName) == "" {
		fmt.Fprintf(os.Stderr, help, "Error: domain is not set.")
		os.Exit(1)
	}
	if os.Getenv(tenantEnvName) == "" {
		fmt.Fprintf(os.Stderr, help, "Error: tenant is not set.")
		os.Exit(1)
	}
	if os.Getenv(usernameEnvName) == "" {
		fmt.Fprintf(os.Stderr, help, "Error: username is not set.")
		os.Exit(1)
	}
	if os.Getenv(passwordEnvName) == "" {
		fmt.Fprintf(os.Stderr, help, "Error: password is not set.")
		os.Exit(1)
	}

	// Binary.
	fmt.Fprintln(os.Stderr, "[TestMain] Compiling test binary.")
	cmd := exec.Command("go", "test", "-c", "-covermode=count", "-coverpkg=./...", "-o=./tests/e2e/dsv.test")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	cmd.Dir = filepath.Join(workDir, "..", "..")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] Error: Failed to compile test binary: %v.\n", err)
		os.Exit(1)
	}
	binPath = "./dsv.test"

	// Temp directory.
	fmt.Fprintln(os.Stderr, "[TestMain] Creating directory for temporary files.")
	tDir, err := os.MkdirTemp("", "cli-config-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] Error: Failed to create temp directory: %v.\n", err)
		os.Exit(1)
	}
	tmpDirPath = tDir
	fmt.Fprintf(os.Stderr, "[TestMain] Temp directory path: %s.\n", tmpDirPath)

	// Coverage directory.
	covDirPath = filepath.Join(workDir, "coverage")
	fmt.Fprintf(os.Stderr, "[TestMain] Coverage directory path: %s.\n", covDirPath)
	fmt.Fprintln(os.Stderr, "[TestMain] Removing directory with old coverage data.")
	err = os.RemoveAll(covDirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] Error: Failed to delete directory with old coverage data: %v.\n", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "[TestMain] Creating directory where new coverage data will be collected.")
	err = os.Mkdir(covDirPath, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] Error: Failed to create directory for coverage data: %v.\n", err)
		os.Exit(1)
	}

	// Run tests.
	var code int
	err = resilienceBefore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] resilienceBefore() failed: %v.\n", err)
		code = 1
	} else {
		code = m.Run()
		err = resilienceAfter()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[TestMain] resilienceAfter() failed: %v.\n", err)
		}
	}

	// Clean up.
	fmt.Fprintln(os.Stderr, "[TestMain] Removing test binary.")
	err = os.Remove(binPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] Error: Failed to delete test binary: %v.\n", err)
	}
	fmt.Fprintf(os.Stderr, "[TestMain] Removing temp directory at path %s.\n", tmpDirPath)
	err = os.RemoveAll(tmpDirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[TestMain] Error: Failed to delete temp directory: %v.\n", err)
	}

	os.Exit(code)
}

type environment struct {
	domain      string
	tenant      string
	username    string
	password    string
	certificate string
	privateKey  string
	tmpDirPath  string
}

func newEnv() *environment {
	e := &environment{
		domain:      os.Getenv(domainEnvName),
		tenant:      os.Getenv(tenantEnvName),
		username:    os.Getenv(usernameEnvName),
		password:    os.Getenv(passwordEnvName),
		certificate: os.Getenv(certificateEnvName),
		privateKey:  os.Getenv(privateKeyEnvName),

		tmpDirPath: tmpDirPath,
	}

	return e
}

type console interface {
	ExpectString(string)
	ExpectEOF()
	SendLine(string)
	Send(string)
	SendKeyEnter()
	SendKeyArrowDown()
}

type consoleWithErrorHandling struct {
	console *expect.Console
	t       *testing.T
}

func (c *consoleWithErrorHandling) ExpectString(s string) {
	c.t.Helper()
	c.t.Logf("Expecting %q", s)
	if buf, err := c.console.ExpectString(s); err != nil {
		c.t.Logf("ExpectString(%q) buffer:\n%s", s, buf)
		c.t.Fatalf("ExpectString(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) ExpectEOF() {
	if buf, err := c.console.ExpectEOF(); err != nil {
		c.t.Helper()
		if strings.Contains(err.Error(), "use of closed file") {
			return // Ignore.
		}
		c.t.Logf("ExpectEOF() buffer:\n%s", buf)
		c.t.Fatalf("ExpectEOF() = %v", err)
	}
}

func (c *consoleWithErrorHandling) SendLine(s string) {
	if _, err := c.console.SendLine(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("SendLine(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) Send(s string) {
	if _, err := c.console.Send(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("Send(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) SendKeyEnter() {
	if _, err := c.console.Send(string(terminal.KeyEnter)); err != nil {
		c.t.Helper()
		c.t.Fatalf("Send(terminal.KeyEnter) = %v", err)
	}
}

func (c *consoleWithErrorHandling) SendKeyArrowDown() {
	if _, err := c.console.Send(string(terminal.KeyArrowDown)); err != nil {
		c.t.Helper()
		c.t.Fatalf("Send(terminal.KeyArrowDown) = %v", err)
	}
}

func prepCmd(t *testing.T, args []string) *exec.Cmd {
	var covName string
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		covName = fmt.Sprintf("%s-%s-%d.out", args[0], args[1], time.Now().UnixNano())
	} else {
		covName = fmt.Sprintf("%s-%d.out", args[0], time.Now().UnixNano())
	}
	covPath := filepath.Join(covDirPath, covName)
	t.Logf("Coverage report: %s", covPath)
	binArgs := append(
		[]string{
			fmt.Sprintf("-test.coverprofile=%s", covPath),
			"-test.timeout=2m",
		},
		args...,
	)
	cmd := exec.Command(binPath, binArgs...)
	cmd.Env = append(os.Environ(), "IS_SYSTEM_TEST=true")
	return cmd
}

func run(t *testing.T, command []string) string {
	t.Helper()
	cmd := prepCmd(t, command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.CombinedOutput() = %v", err)
	}

	out := string(output)

	// Remove lines like this added at the end of the output:
	// 		> PASS
	// 		> coverage: 6.8% of statements in ./...
	// 		>
	return out[:strings.Index(out, `PASS`)]
}

func runWithAuth(t *testing.T, e *environment, command string) string {
	t.Helper()
	args := strings.Split(command, " ")
	args = append(args,
		"--auth-type=password",
		fmt.Sprintf("--auth-username=%s", e.username),
		fmt.Sprintf("--auth-password=%s", e.password),
		fmt.Sprintf("--tenant=%s", e.tenant),
		fmt.Sprintf("--domain=%s", e.domain),
	)
	return run(t, args)
}

func runFlow(t *testing.T, command []string, flow func(c console)) {
	t.Helper()

	pty, tty, err := pseudotty.Open()
	if err != nil {
		t.Fatalf("failed to open pseudotty: %v", err)
	}

	term := vt10x.New(vt10x.WithWriter(tty))
	c, err := expect.NewConsole(
		expect.WithStdin(pty),
		expect.WithStdout(term),
		expect.WithCloser(pty, tty),
		// 15 seconds should be enough even for high API response time.
		expect.WithDefaultTimeout(15*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create console: %v", err)
	}
	defer c.Close()

	cmd := prepCmd(t, command)
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()

	go func() {
		flow(&consoleWithErrorHandling{console: c, t: t})
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() = %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Wait() = %v", err)
	}
}

func createFile(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create(%q) = %v", path, err)
	}
	f.Close()
}

func writeFile(t *testing.T, data []byte, path string) {
	t.Helper()
	err := os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		t.Fatalf("os.WriteFile(%q) = %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = %v", path, err)
	}
	return string(bytes)
}

func deleteFile(t *testing.T, path string) {
	t.Helper()
	err := os.Remove(path)
	if err != nil {
		t.Fatalf("os.Remove(%q) = %v", path, err)
	}
}

func requireLine(t *testing.T, lines string, line string) {
	t.Helper()
	for _, l := range strings.Split(lines, "\n") {
		if strings.TrimSpace(l) == line {
			return
		}
	}
	t.Fatalf("Line %q not found in lines:\n%s", line, lines)
}

func requireContains(t *testing.T, s string, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("String %q not found in:\n%s", substr, s)
	}
}
