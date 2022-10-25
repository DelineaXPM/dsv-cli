package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCryptoManualCmd(t *testing.T) {
	_, err := GetCryptoManualCmd()
	assert.Nil(t, err)
}

func TestGetManualKeyUploadCmd(t *testing.T) {
	_, err := GetManualKeyUploadCmd()
	assert.Nil(t, err)
}

func TestGetManualKeyUpdateCmd(t *testing.T) {
	_, err := GetManualKeyUpdateCmd()
	assert.Nil(t, err)
}

func TestGetManualKeyReadCmd(t *testing.T) {
	_, err := GetManualKeyReadCmd()
	assert.Nil(t, err)
}

func TestGetManualKeyDeleteCmd(t *testing.T) {
	_, err := GetManualKeyDeleteCmd()
	assert.Nil(t, err)
}

func TestGetManualKeyRestoreCmd(t *testing.T) {
	_, err := GetManualKeyRestoreCmd()
	assert.Nil(t, err)
}

func TestGetManualKeyEncryptCmd(t *testing.T) {
	_, err := GetManualKeyEncryptCmd()
	assert.Nil(t, err)
}

func TestGetManualKeyDecryptCmd(t *testing.T) {
	_, err := GetManualKeyDecryptCmd()
	assert.Nil(t, err)
}
