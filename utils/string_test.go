package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToSlice(t *testing.T) {
	f := func(t *testing.T, in string, expected []string) {
		t.Helper()
		actual := StringToSlice(in)
		assert.Equal(t, expected, actual)
	}
	f(t, "", []string{""})
	f(t, "[", []string{""})
	f(t, "]", []string{""})
	f(t, "[]", []string{""})

	f(t, "a", []string{"a"})
	f(t, "a,b", []string{"a", "b"})

	f(t, "[a", []string{"a"})
	f(t, "[a,b", []string{"a", "b"})

	f(t, "a]", []string{"a"})
	f(t, "a,b]", []string{"a", "b"})

	f(t, "[a]", []string{"a"})
	f(t, "[a,b]", []string{"a", "b"})
	f(t, "[a, b]", []string{"a", "b"})
	f(t, `["a", "b"]`, []string{`"a"`, `"b"`})
}
