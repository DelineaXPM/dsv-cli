package cmd_test

import (
	"testing"
	cmd "thy/commands"

	"github.com/stretchr/testify/assert"
)

func TestIsValidProfile(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		profile  string
		expected bool
	}{
		{
			profile:  "profile1",
			expected: true,
		},
		{
			profile:  "profile_name",
			expected: true,
		},
		{
			profile:  "profile-name",
			expected: true,
		},
		{
			profile:  "profile%name",
			expected: true,
		},
		{
			profile:  "profilename'",
			expected: true,
		},
		{
			profile:  "=",
			expected: true,
		},
		{
			profile:  "profile name",
			expected: false,
		},
		{
			profile:  " profilename",
			expected: false,
		},
		{
			profile:  "profilename  ",
			expected: false,
		},
		{
			profile:  `"profilename "`,
			expected: false,
		},
	}

	for _, tC := range testCases {
		assert.Equal(tC.expected, cmd.IsValidProfile(tC.profile), tC.profile)
	}
}
