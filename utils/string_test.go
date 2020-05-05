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
