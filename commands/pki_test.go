package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPkiCmd(t *testing.T) {
	_, err := GetPkiCmd()
	assert.Nil(t, err)
}
