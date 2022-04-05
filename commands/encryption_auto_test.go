package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCryptoCmd(t *testing.T) {
	_, err := GetCryptoCmd()
	assert.Nil(t, err)
}

func TestGetAutoKeyCreateCmd(t *testing.T) {
	_, err := GetAutoKeyCreateCmd()
	assert.Nil(t, err)
}

func TestGetEncryptionRotateCmd(t *testing.T) {
	_, err := GetEncryptionRotateCmd()
	assert.Nil(t, err)
}

func TestGetAutoKeyReadMetadataCmd(t *testing.T) {
	_, err := GetAutoKeyReadMetadataCmd()
	assert.Nil(t, err)
}

func TestGetAutoKeyDeleteCmd(t *testing.T) {
	_, err := GetAutoKeyDeleteCmd()
	assert.Nil(t, err)
}

func TestGetAutoKeyRestoreCmd(t *testing.T) {
	_, err := GetAutoKeyRestoreCmd()
	assert.Nil(t, err)
}

func TestGetEncryptCmd(t *testing.T) {
	_, err := GetEncryptCmd()
	assert.Nil(t, err)
}

func TestGetDecryptCmd(t *testing.T) {
	_, err := GetDecryptCmd()
	assert.Nil(t, err)
}
