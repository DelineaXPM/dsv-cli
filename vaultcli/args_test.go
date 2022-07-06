package vaultcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToFlagName(t *testing.T) {
	assert.Equal(t, "flag1-flag2-flag3", ToFlagName("flag1.flag2-flag3"))
}

func TestGetFlagVal(t *testing.T) {
	const flag = "config"

	f := func(args []string, expected string) {
		t.Helper()
		val := GetFlagVal(flag, args)
		assert.Equal(t, expected, val)
	}

	// Short form.
	f([]string{"-c", "cfg_val", "another_arg"}, "cfg_val")
	f([]string{"-c=cfg_val", "another_arg"}, "cfg_val")

	// Short form. Check if no conflicts with '-c' suffix of the argument value.
	f([]string{"secret-c", "--profile", "testprofile", "-c=cfg_val"}, "cfg_val")
	f([]string{"secret-c", "--profile", "testprofile", "-c", "cfg_val"}, "cfg_val")

	// Long form.
	f([]string{"--config", "cfg_val", "another_arg"}, "cfg_val")
	f([]string{"--config=cfg_val", "another_arg"}, "cfg_val")

	// Long form. Check no conflicts with '--config' suffix of the argument value.
	f([]string{"secret--config", "--profile", "testprofile", "--config", "cfg_val"}, "cfg_val")
	f([]string{"secret--config", "--profile", "testprofile", "--config=cfg_val"}, "cfg_val")
}

func TestGetFilenameFromArgs(t *testing.T) {
	f := func(args []string, expected string) {
		t.Helper()
		actual := GetFilenameFromArgs(args)
		assert.Equal(t, expected, actual)
	}

	// No match.
	f([]string{}, "")
	f([]string{"--path", "pth1"}, "")
	f([]string{"--data", "str1"}, "")
	f([]string{"--data"}, "")
	f([]string{"-d", "str1"}, "")
	f([]string{"-d"}, "")

	// Long flag used.
	f([]string{"--data", "@file1"}, "file1")
	f([]string{"--data", "@file1", "--path", "pth1"}, "file1")
	f([]string{"--path", "pth1", "--data", "@file1"}, "file1")

	// Short flag used.
	f([]string{"-d", "@file1"}, "file1")
	f([]string{"-d", "@file1", "--path", "pth1"}, "file1")
	f([]string{"--path", "pth1", "-d", "@file1"}, "file1")
}
