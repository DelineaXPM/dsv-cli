package vaultcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		internalPath string
		expectedErr  bool
	}{
		{"", true},
		{"a", false},
		{"a:a", false},
		{":a:a", true},
		{"/a:a", true},
		{"secrets:a:a", false},
		{"secrets:a::a", true},
		{":secrets:a:a", true},
		{"/secrets:a:a", true},
		{"/secrets/a:a", true},
		{"secrets/a:a", false},
		{"secrets/1:a", false},
		{"secrets/+:a", true},
		{"secrets/+1:0", false},
		{"secrets:--:a", true},
		{"secrets:+1:-/@@", true},
		{"a$:b", true},
		{"a:b/]", true},
		{"secrets:mari+ia@gmail.com/secret", false},
		{"secrets+:a-:b@/c./d/e", false},
	}
	for _, test := range tests {
		t.Run(test.internalPath, func(t *testing.T) {
			err := ValidatePath(test.internalPath)
			assert.Equal(t, test.expectedErr, err != nil)
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		internalPath string
		expectedErr  bool
	}{
		{"", true},
		{"t", false},
		{"test", false},
		{"te-st", false},
		{"te_st", false},
		{"test_001", false},
		{"123", false},
		{"0", false},
		{"test&", true},
		{"test.test", true},
		{"test+test", true},
		{"[test]", true},
	}
	for _, test := range tests {
		t.Run(test.internalPath, func(t *testing.T) {
			err := ValidateName(test.internalPath)
			assert.Equal(t, test.expectedErr, err != nil)
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		internalPath string
		expectedErr  bool
	}{
		{"", true},
		{"t", false},
		{"test", false},
		{"te-st", false},
		{"te_st", false},
		{"test_001", false},
		{"123", false},
		{"0", false},
		{"test+test", false},
		{"@@@", false},
		{".", false},
		{"+", false},
		{"test@test.com", false},
		{"test&", true},
		{"[test]", true},
		{"te*t", true},
	}
	for _, test := range tests {
		t.Run(test.internalPath, func(t *testing.T) {
			err := ValidateUsername(test.internalPath)
			assert.Equal(t, test.expectedErr, err != nil)
		})
	}
}
