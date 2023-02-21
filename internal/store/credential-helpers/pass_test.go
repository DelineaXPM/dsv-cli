package credhelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPathFromUrl(t *testing.T) {
	testPath := "some-path"
	path := getPathFromURL(testPath)
	assert.Equal(t, path, "thy/some/cGF0aA==")
	url, err := getURLFromPath(path)
	assert.Nil(t, err)
	assert.Equal(t, url, testPath)
}

func TestGetPassDir(t *testing.T) {
	path := getPassDir()
	assert.Contains(t, path, ".password-store")
}
