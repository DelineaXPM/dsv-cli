package credhelpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPathFromUrl(t *testing.T) {
	testPath := "some-path"
	path := getPathFromUrl(testPath)
	assert.Equal(t, path, "thy/some/cGF0aA==")
	url, err := getUrlFromPath(path)
	assert.Nil(t, err)
	assert.Equal(t, url, testPath)
}

func TestGetPassDir(t *testing.T) {
	path := getPassDir()
	assert.Contains(t, path, ".password-store")
}
