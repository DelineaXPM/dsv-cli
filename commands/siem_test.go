package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSiemCmd(t *testing.T) {
	_, err := GetSiemCmd()
	assert.Nil(t, err)
}

func TestGetSiemCreateCmd(t *testing.T) {
	_, err := GetSiemCreateCmd()
	assert.Nil(t, err)
}

func TestGetSiemUpdateCmd(t *testing.T) {
	_, err := GetSiemUpdateCmd()
	assert.Nil(t, err)
}

func TestGetSiemReadCmd(t *testing.T) {
	_, err := GetSiemReadCmd()
	assert.Nil(t, err)
}

func TestGetSiemDeleteCmd(t *testing.T) {
	_, err := GetSiemDeleteCmd()
	assert.Nil(t, err)
}
