package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBYOKCmd(t *testing.T) {
	_, err := GetBYOKCmd()
	assert.Nil(t, err)
}

func TestBYOKUpdateCmd(t *testing.T) {
	_, err := GetBYOKUpdateCmd()
	assert.Nil(t, err)
}
