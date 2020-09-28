package cmd

import (
	"strings"
	"testing"

	cst "thy/constants"
	"thy/utils"

	"github.com/spf13/viper"
)

func TestIsInit(t *testing.T) {
	tests := []struct {
		args     string
		expected bool
	}{
		{"", false},
		{cst.CmdRoot, false},
		{cst.CmdRoot + " " + "secret", false},
		{cst.CmdRoot + " " + "secret init", false},
		{cst.CmdRoot + " " + "cli-config", false},

		{"init", true},
		{"init --dev devbambe.com", true},
		{cst.CmdRoot + " " + "init", true},
		{cst.CmdRoot + " " + "init --dev devbambe.com", true},

		{"cli-config init", true},
		{"cli-config init --profile local", true},
		{cst.CmdRoot + " " + "cli-config init", true},
		{cst.CmdRoot + " " + "cli-config init --profile local", true},
	}

	for _, tc := range tests {
		args := strings.Split(tc.args, " ")
		got := IsInit(args)
		if got != tc.expected {
			t.Errorf("Expected IsInit(%v) to return %v, but got %v", args, tc.expected, got)
		}
	}
}

func TestIsBareArgs(t *testing.T) {
	tests := []struct {
		args     string
		expected bool
	}{
		{"--profile local", true},
		{"--config cfg", true},
		{"--profile local --config cfg", true},
		{"-v", true},
		{"--verbose", true},
		{"--profile local -v", true},
		{"--profile local --config cfg--verbose", true},

		{"--data", false},
		{"--path databases/mongo-db01 --data '{\"Key\":\"Value\"}'", false},
	}

	for _, tc := range tests {
		viper.Reset()
		args := strings.Split(tc.args, " ")
		if utils.Contains(args, "--profile") {
			viper.Set(cst.Profile, "abc")
		}
		if utils.Contains(args, "--config") {
			viper.Set(cst.Config, "cfg")
		}
		got := OnlyGlobalArgs(args)
		if got != tc.expected {
			t.Errorf("Expected OnlyGlobalArgs(%v) to return %v, but got %v", args, tc.expected, got)
		}
	}
}
