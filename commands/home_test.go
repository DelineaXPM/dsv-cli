package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHomeCmd(t *testing.T) {
	_, err := GetHomeCmd()
	assert.Nil(t, err)
}
