package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureNoFmtUsage(t *testing.T) {
	t.Skip("This should be enforced by linting now")
	files, err := getGoFilePaths()
	if err != nil {
		t.Fatalf("Unable to get a list of Go files in the project: %v", err)
	}

	fmtUsage := []byte("fmt.P")

	t.Logf("Processing %d files", len(files))
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("Unable to read the file %s: %v", f, err)
		}
		if bytes.Contains(b, fmtUsage) {
			t.Errorf("Usage of 'fmt' package found in the file '%s'", f)
		}
	}
}

func getGoFilePaths() ([]string, error) { //nolint:unused // leaving for now, until remove test case
	walkRoot := ".." + string(os.PathSeparator)
	privPath := string(os.PathSeparator) + "."

	files := make([]string, 0, 200)
	err := filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, err error) error {
		if strings.Contains(path, privPath) {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "test.go") {
			return nil
		}
		if strings.HasPrefix(path, walkRoot+"cicd") {
			return nil
		}
		if strings.HasPrefix(path, walkRoot+"fake") {
			return nil
		}
		if strings.HasPrefix(path, walkRoot+"vendor") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
