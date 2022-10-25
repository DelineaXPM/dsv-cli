package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	_, err := NewCommand(CommandArgs{})
	assert.Error(t, err)
}
