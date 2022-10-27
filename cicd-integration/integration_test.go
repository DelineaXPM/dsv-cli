package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/utils/test_helpers"

	"github.com/gobuffalo/uuid"
	"golang.org/x/sys/execabs"
)

var update = flag.Bool("update", false, "update golden case files")

var binaryName = constants.CmdRoot

const configPath = "cicd-integration/.thy.yml"

func TestCliArgs(t *testing.T) {
	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("[TestCliArgs] Unable to determine working directory: %v.", err)
	}
	binary := path.Join(workDir, binaryName+".test")

	t.Logf("[TestCliArgs] Working directory: %s", workDir)
	t.Logf("[TestCliArgs] Path to binary: %s", binary)

	err = os.Mkdir("coverage", os.ModeDir)

	// if the error is not nil AND it's not an already exists error
	if err != nil && !os.IsExist(err) {
		t.Fatalf("[TestCliArgs] os.Mkdir(coverage, os.ModeDir): %v", err)
	}
	t.Log("[TestCliArgs] Successfully created the directory for coverage reports.")

	for _, tt := range synchronousCases {
		t.Run(tt.name, func(t *testing.T) {
			outfile := path.Join("coverage", tt.name+"coverage.out")

			args := []string{"-test.coverprofile", outfile}
			args = append(args, tt.args...)
			args = append(args, "--config", configPath)

			cmd := execabs.Command(binary, args...)
			output, _ := cmd.CombinedOutput()

			actual := string(output)
			if strings.LastIndex(actual, "PASS") > -1 {
				actual = actual[:strings.LastIndex(actual, "PASS")]
			}
			if strings.LastIndex(actual, "FAIL") > -1 {
				actual = actual[:strings.LastIndex(actual, "FAIL")]
			}
			actualTrimmed := strings.TrimSpace(actual)

			if *update {
				if tt.output.MatchGoldenCase {
					writeFixture(t, tt.name, []byte(actualTrimmed))
				}
				return
			}

			expected := tt.output.RegexMatch
			if tt.output.MatchGoldenCase {
				expected = loadFixture(t, tt.name)
				expected = regexp.QuoteMeta(expected)
				expected = "^" + expected + "$"
			}

			matcher := regexp.MustCompile(expected)
			match := matcher.MatchString(actualTrimmed)
			if !match {
				t.Fatalf("actual:\n%s,\n expected:\n%s", actualTrimmed, expected)
			}
		})
	}
}

var (
	certPath       = strings.Join([]string{"cicd-integration", "data", "cert.pem"}, string(filepath.Separator))
	privateKeyPath = strings.Join([]string{"cicd-integration", "data", "key.pem"}, string(filepath.Separator))
	csrPath        = strings.Join([]string{"cicd-integration", "data", "csr.pem"}, string(filepath.Separator))
)

const (
	manualKeyPath    = "thekey:first"
	manualPrivateKey = "MnI1dTh4L0E/RChHK0tiUGVTaFZtWXEzczZ2OXkkQiY="
	manualKeyNonce   = "S1NzeHdFcHB6b1Bz"
	plaintext        = "hello there"
	ciphertext       = "8Tns2mbY/w6YHoICfiDGQM+rDlQzwrZWpqK7"
)

func TestMain(m *testing.M) {
	_, err := strconv.ParseBool(os.Getenv("GO_INTEGRATION_TEST"))
	if err != nil {
		fmt.Println("[SKIPPED]: GO_INTEGRATION_TEST must be set to 1/true to run integration tests")
		return
	}

	var rootDir string
	if out, err := execabs.Command("git", "rev-parse", "--show-toplevel").CombinedOutput(); err == nil {
		rootDir = strings.TrimRight(string(out), " \n")
	} else {
		rootDir = "../"
	}

	if err := os.Chdir(rootDir); err != nil {
		log.Fatal(err)
	}

	if err := test_helpers.AddEncryptionKey(os.Getenv("TEST_TENANT"), os.Getenv("USER_NAME"), os.Getenv("DSV_USER_PASSWORD")); err != nil {
		log.Fatalf("could not create encryption key: %v", err)
	}
	makeCmd := execabs.Command("make", "build-test")
	if err := makeCmd.Run(); err != nil {
		log.Fatalf("could not make binary for %s: %v", binaryName, err)
	}

	cert, key, err := generateRootWithPrivateKey()
	csr, err := generateCSR()
	os.WriteFile(certPath, cert, 0o644)
	os.WriteFile(privateKeyPath, key, 0o644)
	os.WriteFile(csrPath, csr, 0o644)

	defer os.Remove(certPath)
	defer os.Remove(privateKeyPath)
	defer os.Remove(csrPath)

	// Before and after *all* tests, make sure any modifications to the config are reverted.
	// Reading and writing the config before and after *each* test is not feasible, as there may be tests that
	// intentionally modify the config to test for presence or absence of a property or modification of a value.
	config, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("could not read config: %v", err)
		os.Exit(1)
	}
	_ = os.Setenv("IS_SYSTEM_TEST", "true")
	m.Run()
	_ = os.Unsetenv("IS_SYSTEM_TEST")

	err = os.WriteFile(configPath, config, 0o644)
	if err != nil {
		fmt.Printf("could not write config: %v", err)
		os.Exit(1)
	}
}

type outputValidation struct {
	RegexMatch      string
	MatchGoldenCase bool
}

func outputIs(exactMatch string) outputValidation {
	return outputValidation{
		RegexMatch: "^" + regexp.QuoteMeta(exactMatch) + "$",
	}
}

func outputPattern(regex string) outputValidation {
	return outputValidation{
		RegexMatch: regex,
	}
}

func outputEmpty() outputValidation {
	return outputValidation{
		RegexMatch: "^$",
	}
}

func outputGolden() outputValidation {
	return outputValidation{
		MatchGoldenCase: true,
	}
}

var (
	secret1Name       string
	secret1Desc       string
	secret1Tag        string
	secret1Data       string
	secret1Attributes string
	secret1DataFmt    string
)

var (
	adminUser, adminPass    string
	user1, user1Pass        string
	roleName                string
	policyName, policy2Name string
	groupName               string
	authProvider            string
	synchronousCases        []struct {
		name   string
		args   []string
		output outputValidation
	}
)

func init() {
	if err := generateThyYml(".thy.yml.template", ".thy.yml"); err != nil {
		panic(err)
	}

	if err := generateThyYml("data/policy.json", "data/test_policy.json"); err != nil {
		panic(err)
	}

	u, _ := uuid.NewV4()
	t, _ := uuid.NewV4()
	adminUser = os.Getenv("ADMIN_USER")
	adminPass = os.Getenv("DSV_ADMIN_PASS")
	secret1Name = u.String() + "z" // Avoid UUID detection on the API side.
	secret1Tag = t.String()
	secret1Desc = `desc of s1`
	secret1Data = `{"field":"secret password"}`
	secret1Attributes = fmt.Sprintf(`{"tag":"%s", "tll": 100}`, secret1Tag)
	secret1DataFmt = `"field": "secret password"`

	r, _ := uuid.NewV4()
	roleName = r.String()

	user1 = os.Getenv("USER1_NAME")
	user1Pass = os.Getenv("DSV_USER1_PASSWORD")
	groupName = u.String()

	policyName = "secrets:" + secret1Name
	p2, _ := uuid.NewV4()
	policy2Name = "secrets:servers:" + p2.String()
	policy2File := strings.Join([]string{"cicd-integration", "data", "test_policy.json"}, string(filepath.Separator))

	homeSecretPath := "sweet/home/secret1"

	existingRootSecret := "existingRoot"
	certStoreSecret := "myroot"
	leafSecretPath := "myleaf"

	ap, _ := uuid.NewV4()
	authProvider = "aws-" + ap.String()

	today := time.Now()
	monthAgoDate := today.AddDate(0, 0, -30).Format("2006-01-02")
	futureDate := today.AddDate(0, 0, 7).Format("2006-01-02")

	synchronousCases = []struct {
		name   string
		args   []string
		output outputValidation
	}{
		// secret operations
		// TODO investigate test setup, as the order of calls matters for some reason.
		{"secret-create-1-pass", []string{"secret", "create", "--path", secret1Name, "--data", secret1Data, "--attributes", secret1Attributes, "--desc", secret1Desc, "-f", ".data", "-v"}, outputPattern(secret1DataFmt)},
		{"secret-update-pass", []string{"secret", "update", "--path", secret1Name, "--desc", "updated secret", "-f", ".data", "-v"}, outputPattern(secret1DataFmt)},
		{"secret-rollback-pass", []string{"secret", "rollback", "--path", secret1Name, "-f", ".data"}, outputPattern(secret1DataFmt)},
		// {"secret-search-find-pass", []string{"secret", "search", secret1Name[:3], "'data.[0].name'"}, outputPattern(secret1Name)},
		{"secret-search-tags", []string{"secret", "search", secret1Tag, "--search-field", "attributes.tag"}, outputPattern(secret1Name)},
		{"secret-create-fail-dup", []string{"secret", "create", "--path", secret1Name, "--data", secret1Data, "", ".message"}, outputPattern(`"message": "error creating secret, secret at path already exists"`)},
		{"secret-describe-1-pass", []string{"secret", "describe", "--path", secret1Name, "-f", ".description"}, outputIs(secret1Desc)},
		{"secret-read-1-pass", []string{"secret", "read", "--path", secret1Name, "-f", ".data"}, outputPattern(secret1DataFmt)},
		{"secret-read-implicit-pass", []string{"secret", secret1Name, "-f", ".data"}, outputPattern(secret1DataFmt)},
		{"secret-search-none-pass", []string{"secret", "search", "hjkl"}, outputPattern(`"data": \[\]`)},
		{"secret-soft-delete", []string{"secret", "delete", secret1Name}, outputPattern("will be removed")},
		{"secret-read-fail", []string{"secret", "read", secret1Name}, outputPattern("will be removed")},
		{"secret-restore", []string{"secret", "restore", secret1Name}, outputEmpty()},

		// policy operations
		{"policy-help", []string{"policy", ""}, outputPattern(`Execute an action on a policy.*`)},
		{"policy-create-pass", []string{"policy", "create", "--path", policyName, "--resources", policyName, "--actions", "read", "--subjects", "users:" + user1}, outputPattern(fmt.Sprintf(`"path":\s*"%s"`, policyName))},
		{"policy-create-file-pass", []string{"policy", "create", "--path", policy2Name, "--data", "@" + policy2File}, outputPattern(fmt.Sprintf(`"path":\s*"%s"`, policy2Name))},
		{"policy-read-pass", []string{"policy", "read", "--path", policyName}, outputPattern(fmt.Sprintf(`"path":\s*"%s"`, policyName))},
		{"policy-search-pass", []string{"policy", "search", "--query", policyName}, outputPattern(fmt.Sprintf(`"path":\s*"%s"`, policyName))},
		{"policy-update-pass", []string{"policy", "update", "--path", policyName, "--resources", policyName, "--actions", "read,delete", "--subjects", "users:" + user1}, outputPattern(`"delete"`)},
		{"policy-rollback-pass", []string{"policy", "rollback", "--path", policyName}, outputPattern(fmt.Sprintf(`"path":\s*"%s"`, policyName))},
		{"policy-soft-delete", []string{"policy", "delete", policyName}, outputPattern("will be removed")},
		{"policy-read-fail", []string{"policy", "read", policyName}, outputPattern("will be removed")},
		{"policy-restore", []string{"policy", "restore", policyName}, outputEmpty()},

		// auth provider operations
		{"auth-provider-help", []string{"config", "auth-provider", ""}, outputPattern(`Execute an action on an auth-provider.*`)},
		{"auth-provider-create", []string{"config", "auth-provider", "create", "--name", authProvider, "--type", "aws", "--aws-account-id", "1234"}, outputPattern(fmt.Sprintf(`"name":\s*"%s"`, authProvider))},
		{"auth-provider-read", []string{"config", "auth-provider", "read", authProvider}, outputPattern(fmt.Sprintf(`"name":\s*"%s"`, authProvider))},
		{"auth-provider-update", []string{"config", "auth-provider", "update", "--name", authProvider, "--type", "aws", "--aws-account-id", "65433"}, outputPattern(fmt.Sprintf(`"accountId":\s*"%s"`, "65433"))},
		{"auth-provider-rollback-pass", []string{"config", "auth-provider", "rollback", "--name", authProvider}, outputPattern(fmt.Sprintf(`"accountId":\s*"%s"`, "1234"))},

		// user operations
		{"user-help", []string{"user", ""}, outputPattern(`Execute an action on a user.*`)},
		{"user-create-pass", []string{"user", "create", "--username", user1, "--password", user1Pass}, outputPattern(`"userName": "mrmittens"`)},
		{"user-read-pass", []string{"user", "read", user1}, outputPattern(`"userName": "mrmittens"`)},
		{"user-read-implicit-pass", []string{"user", user1}, outputPattern(`"userName": "mrmittens"`)},
		{"user-create-fail", []string{"user", "create", "--username", user1, "--password", user1Pass}, outputPattern(`"code": 400`)},
		// {"user-search-find-pass", []string{"user", "search", user1[:3], "-f", "'data.[0].userName'"}, outputPattern(user1)},
		{"user-search-none-pass", []string{"user", "search", "erkjwr"}, outputPattern(`"data": \[\]`)},
		{"user-soft-delete", []string{"user", "delete", user1}, outputPattern("will be removed")},
		{"user-read-fail", []string{"user", "read", user1}, outputPattern("will be removed")},
		{"user-restore", []string{"user", "restore", user1}, outputEmpty()},
		{"user-read-pass-verify-restore", []string{"user", "read", user1}, outputPattern(`"userName": "mrmittens"`)},
		{"user-create-provider-missing", []string{"user", "create", "--username", "bob", "--external-id", "1234"}, outputPattern("provider is required")},
		{"user-create-external-id-missing", []string{"user", "create", "--username", "bob", "--provider", authProvider}, outputPattern("external ID is required")},

		// group operations
		{"group-help", []string{"group", ""}, outputPattern(`Execute an action on a group.*`)},
		{"group-create-pass", []string{"group", "create", "--group-name", groupName, "--members", user1}, outputPattern(`.*` + "\"errors\": {}" + `.*`)},
		{"group-read-pass", []string{"group", "read", groupName}, outputPattern(groupName)},
		{"group-delete-member-pass", []string{"group", "delete-members", "--group-name", groupName, "--members", user1}, outputEmpty()},
		{"group-read-pass", []string{"group", "read", groupName}, outputPattern(groupName)},
		{"group-soft-delete", []string{"group", "delete", groupName}, outputPattern("will be removed")},
		{"group-read-fail", []string{"group", "read", groupName}, outputPattern("will be removed")},
		{"group-restore", []string{"group", "restore", groupName}, outputEmpty()},

		// role operations
		{"role-create-pass", []string{"role", "create", "--name", roleName}, outputPattern(fmt.Sprintf(`"name":\s*"%s"`, roleName))},

		// client operations
		{"client-create-pass", []string{"client", "create", "--role", roleName}, outputPattern(`"role":\s*"[^"]+"`)},
		{"client-create-fail", []string{"client", "create", "--role", roleName + "doesntexist"}, outputPattern(`"code": 400`)},
		{"client-search-pass", []string{"client", "search", "--role", roleName}, outputPattern(`"clientId"`)},

		// delegated access operations
		{"user-auth-pass", []string{"auth", "-u", user1, "-p", user1Pass}, outputPattern(`"accessToken":\s*"[^"]+",\s*"expiresIn"`)},
		{"user-auth-pass-failed", []string{"auth", "-u", user1, "-p", "user1fail"}, outputPattern(`{"code":401,"message":"unable to authenticate"}`)},
		{"user-access-pass", []string{"secret", "read", secret1Name, "-u", user1, "-p", user1Pass}, outputPattern(secret1DataFmt)},
		{"user-access-fail-action", []string{"secret", "update", secret1Name, "-u", user1, "-p", user1Pass, "-d", `{"field":"updated secret 1"}`}, outputPattern("Invalid permissions")},
		{"user-access-fail-resource", []string{"secret", "read", "secret-idonthavepermissionon", "-u", user1, "-p", user1Pass, "-f", ".data"}, outputPattern("Invalid permissions")},

		// cli-config operations
		{"cli-config-help", []string{"cli-config", ""}, outputPattern(`Execute an action on the cli config.*`)},
		{"cli-config-read-pass", []string{"cli-config", "read"}, outputGolden()},
		// Force update to the config with the same correct password.
		{"cli-config-update-pass", []string{"cli-config", "update", "auth.password", adminPass}, outputEmpty()},

		// Make sure config now has a `securePassword` key.
		{"cli-config-read-2-pass", []string{"cli-config", "read"}, outputPattern(`securePassword`)},

		// Config will not be written, if auth fails upon password update.
		{"cli-config-update-fail", []string{"cli-config", "update", "auth.password", "wrong-password"}, outputPattern(`Please check your credentials and try again.`)},

		{"token-clear-pass", []string{"auth", "clear"}, outputEmpty()},
		{"user-auth-success", []string{"auth"}, outputPattern(`accessToken`)},

		{"cli-config-add-key", []string{"cli-config", "update", "key", "value"}, outputEmpty()},
		{"cli-config-remove-key", []string{"cli-config", "update", "key", "0"}, outputEmpty()},

		{"cli-config-update-2-pass", []string{"cli-config", "update", "auth.password", adminPass}, outputEmpty()},

		{"usage-pass", []string{"usage", "--startdate", monthAgoDate}, outputPattern("requestsUsed")},
		{"usage-fail", []string{"usage", "--startdate", futureDate}, outputPattern("Usage start date cannot be in the future")},

		// {"logs-pass", []string{"logs", "--startdate", monthAgoDate}, outputPattern("data")},
		// {"audit-pass", []string{"audit", "--startdate", monthAgoDate}, outputPattern("data")},

		// config operations
		{"config-help", []string{"config", "--help"}, outputPattern(`Execute an action on the.*`)},
		{"config-get-implicit-pass", []string{"config"}, outputPattern(`"permissionDocument":`)},
		{"config-get-pass", []string{"config", "read"}, outputPattern(`"permissionDocument":`)},

		// Home
		{"home-help", []string{"home", ""}, outputPattern(`Work with secrets in a personal user space`)},
		{"home-create", []string{"home", "create", homeSecretPath, "--desc", "some description"}, outputPattern(strings.ReplaceAll(homeSecretPath, "/", ":"))},
		{"home-update", []string{"home", "update", homeSecretPath, "--desc", "updated description"}, outputPattern(strings.ReplaceAll(homeSecretPath, "/", ":"))},
		{"home-read", []string{"home", "read", homeSecretPath}, outputPattern("updated description")},
		{"home-delete", []string{"home", "delete", homeSecretPath}, outputPattern("secret marked for deletion")},
		{"home-restore", []string{"home", "restore", homeSecretPath}, outputEmpty()},
		{"home-rollback", []string{"home", "rollback", homeSecretPath}, outputPattern(`"version": "2"`)},
		{"home-get-by-version", []string{"home", "read", homeSecretPath, "version", "2"}, outputPattern(`"version": "2"`)},

		// EaaS-Manual
		{"crypto-manual-key-upload", []string{"crypto", "manual", "key-upload", "--path", manualKeyPath, "--private-key", manualPrivateKey, "--nonce", manualKeyNonce, "--scheme", "symmetric"}, outputPattern(`"version": "0"`)},
		{"crypto-manual-key-read", []string{"crypto", "manual", "key-read", "--path", manualKeyPath}, outputPattern(`"version": "0"`)},
		{"crypto-manual-encrypt", []string{"crypto", "manual", "encrypt", "--path", manualKeyPath, "--data", plaintext}, outputPattern(`"version": "0"`)},
		{"crypto-manual-decrypt", []string{"crypto", "manual", "decrypt", "--path", manualKeyPath, "--data", ciphertext}, outputPattern(`"data": "hello there"`)},
		{"crypto-manual-key-update", []string{"crypto", "manual", "key-update", "--path", manualKeyPath, "--private-key", manualPrivateKey}, outputPattern(`"version": "1"`)},

		// PKI
		{
			"register-root-cert",
			[]string{
				"pki", "register", "--rootcapath", existingRootSecret,
				"--certpath", "@" + certPath, "--privkeypath", "@" + privateKeyPath, "--domains", leafCommonName, "--maxttl", "250h",
			},
			outputPattern("certificate"),
		},

		{
			"sign-with-root-cert",
			[]string{
				"pki", "sign", "--rootcapath", existingRootSecret,
				"--csrpath", "@" + csrPath, "--ttl", "100H",
			},
			outputPattern("certificate"),
		},

		{
			"generate-root-cert",
			[]string{
				"pki", "generate-root", "--rootcapath", certStoreSecret,
				"--domains", leafCommonName, "--common-name", "thycotic.com", "--maxttl", "60d",
			},
			outputPattern("certificate"),
		},

		{
			"generate-leaf-cert",
			[]string{
				"pki", "leaf", "--rootcapath", certStoreSecret,
				"--common-name", leafCommonName, "--ttl", "5D", "--store-path", leafSecretPath,
			},
			outputPattern("certificate"),
		},

		{
			"generate-ssh-cert",
			[]string{
				"pki", "ssh-cert", "--rootcapath", certStoreSecret, "--leafcapath",
				leafSecretPath, "--principals", "root,ubuntu", "--ttl", "52w",
			},
			outputPattern("sshCertificate"),
		},

		{"user-update-pass", []string{"user", "update", "--username", user1, "--password", "New_password@2"}, outputPattern(`"userName": "mrmittens"`)},

		// cleanup
		{"secret-delete-1-pass", []string{"secret", "delete", secret1Name, "--force"}, outputEmpty()},
		{"user-delete", []string{"user", "delete", user1, "--force"}, outputEmpty()},
		{"role-delete-fail", []string{"role", "delete", roleName, "--force"}, outputPattern(`cannot delete role`)},
		{"auth-provider-delete", []string{"config", "auth-provider", "delete", "--name", authProvider, "--force"}, outputEmpty()},
		{"policy-delete", []string{"policy", "delete", "--path", policyName, "--force"}, outputEmpty()},
		{"policy2-delete", []string{"policy", "delete", "--path", policy2Name, "--force"}, outputEmpty()},
		{"cert-secret-delete", []string{"secret", "delete", "--path", certStoreSecret, "--force"}, outputEmpty()},
		{"rootCA-secret-delete", []string{"secret", "delete", "--path", existingRootSecret, "--force"}, outputEmpty()},
		{"leafCA-secret-delete", []string{"secret", "delete", "--path", leafSecretPath, "--force"}, outputEmpty()},
		{"home-secret-delete", []string{"home", "delete", homeSecretPath, "--force"}, outputEmpty()},
		{"crypto-manual-key-delete", []string{"crypto", "manual", "key-delete", "--path", manualKeyPath, "--force"}, outputEmpty()},
	}
}

func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}
	return filepath.Join(filepath.Dir(filename), "cases", fixture)
}

func writeFixture(t *testing.T, fixture string, content []byte) {
	err := os.WriteFile(fixturePath(t, fixture), content, 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

func loadFixture(t *testing.T, fixture string) string {
	tmpl, err := template.ParseFiles(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	var tmplBytes bytes.Buffer
	err = tmpl.Execute(&tmplBytes, envToMap())
	if err != nil {
		t.Fatal(err)
	}
	return tmplBytes.String()
}

func generateThyYml(inPath, outPath string) error {
	t, err := template.ParseFiles(inPath)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return t.Execute(outFile, envToMap())
}

func envToMap() map[string]string {
	evpMap := map[string]string{}

	for _, v := range os.Environ() {
		split := strings.Split(v, "=")
		if len(split) == 2 {
			evpMap[split[0]] = split[1]
		}
	}
	return evpMap
}
