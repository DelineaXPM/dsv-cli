package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCliConfigCmd(t *testing.T) {
	_, err := GetCliConfigCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigInitCmd(t *testing.T) {
	_, err := GetCliConfigInitCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigUpdateCmd(t *testing.T) {
	_, err := GetCliConfigUpdateCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigClearCmd(t *testing.T) {
	_, err := GetCliConfigClearCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigReadCmd(t *testing.T) {
	_, err := GetCliConfigReadCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigEditCmd(t *testing.T) {
	_, err := GetCliConfigEditCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigUseProfileCmd(t *testing.T) {
	_, err := GetCliConfigUseProfileCmd()
	assert.Nil(t, err)
}
