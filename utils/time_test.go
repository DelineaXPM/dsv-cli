package utils

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input         string
		output        int
		errorExpected bool
	}{
		{"", 0, false},
		{"0", 0, false},
		{"5", 5, false},
		{"1h", 1, false},
		{"100H", 100, false},
		{"3d", 72, false},
		{"61D", 1464, false},
		{"2w", 336, false},
		{"52W", 8736, false},

		{"hello", 0, true},
		{"h23", 0, true},
		{"4a", 0, true},
		{"4.3", 0, true},
		{"4U", 0, true},
		{"M42", 0, true},
		{"3ww", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			d, err := ParseHours(test.input)
			if (err == nil && test.errorExpected) || (err != nil && !test.errorExpected) {
				t.Fatalf("Unexpected error to be %v, but got %v", test.errorExpected, err)
			}
			if d != test.output {
				t.Errorf("Expected to get %d, but got %d", test.output, d)
			}
		})
	}
}
