package cmd

import (
	"strings"
	"testing"

	cst "thy/constants"
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

func TestIsInstall(t *testing.T) {
	tests := []struct {
		args     string
		expected bool
	}{
		{"", false},
		{cst.CmdRoot, false},
		{cst.CmdRoot + " " + "secret", false},
		{cst.CmdRoot + " " + "secret init", false},
		{cst.CmdRoot + " " + "cli-config", false},

		{"--install", true},
		{"-install", true},
		{cst.CmdRoot + " " + "--install", true},
		{cst.CmdRoot + " " + "-install", true},
	}

	for _, tc := range tests {
		args := strings.Split(tc.args, " ")
		got := IsInstall(args)
		if got != tc.expected {
			t.Errorf("Expected IsInstall(%v) to return %v, but got %v", args, tc.expected, got)
		}
	}
}

func TestOnlyGlobalArgs(t *testing.T) {
	allGlobals := func(flags string) {
		t.Helper()
		result := OnlyGlobalArgs(strings.Split(flags, " "))
		if !result {
			t.Errorf("OnlyGlobalArgs(%v) must return true", flags)
		}
	}
	notGlobals := func(flags string) {
		t.Helper()
		result := OnlyGlobalArgs(strings.Split(flags, " "))
		if result {
			t.Errorf("OnlyGlobalArgs(%v) must return false", flags)
		}
	}

	allGlobals("--profile local")
	allGlobals("--profile=local")
	allGlobals("--config cfg")
	allGlobals("--profile local --config cfg")
	allGlobals("-v")
	allGlobals("--verbose")
	allGlobals("--profile local -v")
	allGlobals("--profile local --config cfg--verbose")
	allGlobals("--auth-type password -v")
	allGlobals("--auth-type password --auth-username tom --auth-password r1ddle")
	allGlobals("--auth-type password --auth-username=tom --auth-password r1ddle")
	allGlobals("-a password -u tom -p r1ddle")

	notGlobals("--data")
	notGlobals("--path databases/mongo-db01 --data '{\"Key\":\"Value\"}'")
	notGlobals("--effect allow --auth-type password --auth-username tom --auth-password r1ddle")
	notGlobals("--effect allow -a password -u tom -p r1ddle")
}
