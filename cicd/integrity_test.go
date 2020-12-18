package cicd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureNoFmtUsage(t *testing.T) {
	fmtPackageName := `fmt.P`
	files, err := getListGoFilePaths()
	assert.Nil(t, err)
	for _, f := range files {
		if strings.HasSuffix("test.go", f) || strings.HasSuffix("format.go", f) {
			// dont care about tests, or format package
			continue
		}
		b, err := ioutil.ReadFile(f)
		assert.Nil(t, err)
		assert.False(t, bytes.Contains(b, []byte(fmtPackageName)), fmt.Sprintf(`Usage of "fmt" found in file '%s'`, f))
	}
}
func getListGoFilePaths() ([]string, error) {
	var pathDelim string
	if runtime.GOOS == "windows" {
		pathDelim = "\\"
	} else {
		pathDelim = "/"
	}
	var filesAll []string
	filesGo := make([]string, 0, 500)
	err := filepath.Walk(".."+pathDelim, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, pathDelim+".") {
			return nil
		}
		filesAll = append(filesAll, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, file := range filesAll {
		if !strings.HasSuffix(file, ".go") {
			continue
		}
		if strings.HasPrefix(file, ".."+pathDelim+"cicd") {
			continue
		}
		if strings.HasPrefix(file, ".."+pathDelim+"vendor") {
			continue
		}
		filesGo = append(filesGo, file)
	}
	return filesGo, nil
}
