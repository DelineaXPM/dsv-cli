package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPkiCmd(t *testing.T) {
	_, err := GetPkiCmd()
	assert.Nil(t, err)
}

func TestGetPkiRegisterCmd(t *testing.T) {
	_, err := GetPkiRegisterCmd()
	assert.Nil(t, err)
}

func TestGetPkiSignCmd(t *testing.T) {
	_, err := GetPkiSignCmd()
	assert.Nil(t, err)
}

func TestGetPkiLeafCmd(t *testing.T) {
	_, err := GetPkiLeafCmd()
	assert.Nil(t, err)
}

func TestGetPkiGenerateRootCmd(t *testing.T) {
	_, err := GetPkiGenerateRootCmd()
	assert.Nil(t, err)
}

func TestGetPkiSSHCertCmd(t *testing.T) {
	_, err := GetPkiSSHCertCmd()
	assert.Nil(t, err)
}
