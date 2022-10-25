package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHomeCmd(t *testing.T) {
	_, err := GetHomeCmd()
	assert.Nil(t, err)
}

func TestGetHomeReadCmd(t *testing.T) {
	_, err := GetHomeReadCmd()
	assert.Nil(t, err)
}

func TestGetHomeCreateCmd(t *testing.T) {
	_, err := GetHomeCreateCmd()
	assert.Nil(t, err)
}

func TestGetHomeDeleteCmd(t *testing.T) {
	_, err := GetHomeDeleteCmd()
	assert.Nil(t, err)
}

func TestGetHomeRestoreCmd(t *testing.T) {
	_, err := GetHomeRestoreCmd()
	assert.Nil(t, err)
}

func TestGetHomeUpdateCmd(t *testing.T) {
	_, err := GetHomeUpdateCmd()
	assert.Nil(t, err)
}

func TestGetHomeRollbackCmd(t *testing.T) {
	_, err := GetHomeRollbackCmd()
	assert.Nil(t, err)
}

func TestGetHomeSearchCmd(t *testing.T) {
	_, err := GetHomeSearchCmd()
	assert.Nil(t, err)
}

func TestGetHomeDescribeCmd(t *testing.T) {
	_, err := GetHomeDescribeCmd()
	assert.Nil(t, err)
}

func TestGetHomeEditCmd(t *testing.T) {
	_, err := GetHomeEditCmd()
	assert.Nil(t, err)
}
