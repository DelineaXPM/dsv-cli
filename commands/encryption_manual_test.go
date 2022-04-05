package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCryptoManualCmd(t *testing.T) {
	_, err := GetCryptoManualCmd()
	assert.Nil(t, err)
}
