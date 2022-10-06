package vaultcli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRawConfig_emptyFile(t *testing.T) {
	version, defaultProfile, profiles, err := parseRawConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error parsing <nil> config: %v", err)
	}
	if version != "" {
		t.Fatalf("unexpected version, want <empty string>, got %s", version)
	}

	if defaultProfile != "" {
		t.Fatalf("unexpected default profile, want <empty string>, got %s", defaultProfile)
	}

	if len(profiles) != 0 {
		t.Fatalf("unexpected length of profiles, want 0, got %d", len(profiles))
	}

	version, defaultProfile, profiles, err = parseRawConfig([]byte(``))
	if err != nil {
		t.Fatalf("unexpected error parsing empty config: %v", err)
	}
	if version != "" {
		t.Fatalf("unexpected version, want <empty string>, got %s", version)
	}

	if defaultProfile != "" {
		t.Fatalf("unexpected default profile, want <empty string>, got %s", defaultProfile)
	}

	if len(profiles) != 0 {
		t.Fatalf("unexpected length of profiles, want 0, got %d", len(profiles))
	}
}

func TestParseRawConfig_parsingIssues(t *testing.T) {
	_, _, _, err := parseRawConfig([]byte(`aa`))
	if err == nil {
		t.Fatalf("error should not be nil if parsing invalid YAML config")
	}
}

func TestParseRawConfig_unsupportedVersion(t *testing.T) {
	_, _, _, err := parseRawConfig([]byte(`version: v222
defaultProfile: p1
profiles:
    p1:
        example: example
`))
	if err == nil {
		t.Fatalf("error should not be nil if version is not supported")
	}
}

func TestParseRawConfig_v2Format(t *testing.T) {
	version, defaultProfile, profiles, err := parseRawConfig([]byte(`version: v2
defaultProfile: p1
profiles:
    p1:
        example: example
    p2:
        example: example
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "v2" {
		t.Fatalf("unexpected version, want v2, got %s", version)
	}

	if defaultProfile != "p1" {
		t.Fatalf("unexpected default profile, want p1, got %s", defaultProfile)
	}

	if len(profiles) != 2 {
		t.Fatalf("unexpected length of profiles, want 2, got %d", len(profiles))
	}
}

func TestParseRawConfig_v2MissingDefault(t *testing.T) {
	_, _, _, err := parseRawConfig([]byte(`version: v2
defaultProfile: p3
profiles:
    p1:
        example: example
`))
	if err == nil {
		t.Fatalf("error should not be nil if default profile is missing in profiles")
	}
}

func TestParseRawConfig_v1Format(t *testing.T) {
	version, defaultProfile, profiles, err := parseRawConfig([]byte(`default:
    example: example
p2:
    example: example
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "v1" {
		t.Fatalf("unexpected version, want v1, got %s", version)
	}

	if defaultProfile != "default" {
		t.Fatalf("unexpected default profile, want default, got %s", defaultProfile)
	}

	if len(profiles) != 2 {
		t.Fatalf("unexpected length of profiles, want 2, got %d", len(profiles))
	}
}

func TestParseRawConfig_v1MissingDefault(t *testing.T) {
	_, _, _, err := parseRawConfig([]byte(`p1:
    example: example
`))
	if err == nil {
		t.Fatalf("error should not be nil if default profile is missing in profiles")
	}
}

func TestLookupConfigPath(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("Cannot create temp dir: %v", err)
	}
	thyFile := filepath.Join(tmpdir, ".thy.yml")
	dsvFile := filepath.Join(tmpdir, ".dsv.yml")

	t.Logf("Temp dir: %v", tmpdir)

	// Case 1: Empty directory.
	path := LookupConfigPath(tmpdir)
	if path != dsvFile {
		t.Fatalf("Unexpected result for empty directory. Got path %q", path)
	}

	// Case 2: Directory with only ".thy.yml" file.
	createFile(t, thyFile)
	path = LookupConfigPath(tmpdir)
	if path != thyFile {
		t.Fatalf("Unexpected result for directory with '.thy.yml' file. Got path %q", path)
	}

	// Case 3: Directory with both ".thy.yml" and ".dsv.yml" files.
	createFile(t, dsvFile)
	path = LookupConfigPath(tmpdir)
	if path != dsvFile {
		t.Fatalf("Unexpected result for directory with both '.thy.yml' and 'dsv.yml' files. Got path %q", path)
	}

	// Case 4: Directory with only ".dsv.yml" file.
	deleteFile(t, thyFile)
	path = LookupConfigPath(tmpdir)
	if path != dsvFile {
		t.Fatalf("Unexpected result for directory with '.thy.yml' file. Got path %q", path)
	}

	// Cleanup.
	err = os.RemoveAll(tmpdir)
	if err != nil {
		t.Fatalf("Failed to remove temp dir at path %s: %v", tmpdir, err)
	}
}

func TestGetProfile(t *testing.T) {
	cf, err := NewConfigFile("random-path")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	prf, ok := cf.GetProfile("n1")
	if prf != nil {
		t.Fatal("Expected <nil> profile.")
	}
	if ok {
		t.Fatal("Expected <false> as second returned value.")
	}

	cf.SetProfile(&Profile{Name: "n1", data: make(map[string]interface{})})

	prf, ok = cf.GetProfile("n1")
	if prf == nil {
		t.Fatal("Expected not <nil> profile.")
	}
	if !ok {
		t.Fatal("Expected <true> as second returned value.")
	}
}

func TestListProfiles(t *testing.T) {
	cf, err := NewConfigFile("random-path")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	prfs := cf.ListProfiles()
	if len(prfs) != 0 {
		t.Fatalf("Want len 0, got len %d", len(prfs))
	}

	cf.SetProfile(&Profile{Name: "a2", data: make(map[string]interface{})})
	cf.SetProfile(&Profile{Name: "b1", data: make(map[string]interface{})})
	cf.SetProfile(&Profile{Name: "a1", data: make(map[string]interface{})})

	prfs = cf.ListProfiles()
	if len(prfs) != 3 {
		t.Fatalf("Want len: 3, got: %d", len(prfs))
	}
	if prfs[0].Name != "a1" {
		t.Fatalf("Want first element a1, got: %s", prfs[0].Name)
	}
	if prfs[1].Name != "a2" {
		t.Fatalf("Want second element a2, got: %s", prfs[1].Name)
	}
	if prfs[2].Name != "b1" {
		t.Fatalf("Want third element b1, got: %s", prfs[2].Name)
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

func deleteFile(t *testing.T, path string) {
	t.Helper()
	err := os.Remove(path)
	if err != nil {
		t.Fatalf("os.Remove(%q) = %v", path, err)
	}
}
