package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Regular string",
			input:    "path",
			expected: []string{"path"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "String with only brackets",
			input:    "[]",
			expected: []string{""},
		},
		{
			name:     "String with opening bracket",
			input:    "[path",
			expected: []string{"path"},
		},
		{
			name:     "String with closing bracket",
			input:    "path]",
			expected: []string{"path"},
		},
		{
			name:     "String with both brackets",
			input:    "[path]",
			expected: []string{"path"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := StringToSlice(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestCheckPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefixes []string
		expected bool
	}{
		{
			name:     "One prefix - match",
			input:    "users:json/test/test1/test2",
			prefixes: []string{"users:"},
			expected: true,
		},
		{
			name:     "Two prefix - match",
			input:    "roles:json/test/test1/test2",
			prefixes: []string{"users:", "roles:"},
			expected: true,
		},
		{
			name:     "Not match",
			input:    "users:json/test/test1/test2",
			prefixes: []string{"other:"},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := CheckPrefix(test.input, test.prefixes...)
			assert.Equal(t, test.expected, actual)
		})
	}
}
