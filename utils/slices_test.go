package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlices(t *testing.T) {

	tests := []struct {
		name     string
		input1   []byte
		input2   []byte
		expected bool
	}{
		{
			name:     "Nil_Path",
			expected: true,
		},
		{
			name:     "Input1_Path",
			input1:   []byte(``),
			expected: false,
		},
		{
			name:     "Input2_Path",
			input2:   []byte(`val`),
			expected: false,
		},
		{
			name:     "Input2_Path",
			input1:   []byte(`al`),
			input2:   []byte(`val`),
			expected: false,
		},
		{
			name:     "Happy_Path",
			input1:   []byte(`val`),
			input2:   []byte(`lav`),
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			actual := SlicesEqual(test.input1, test.input2)
			assert.Equal(t, test.expected, actual)

		})
	}
}
