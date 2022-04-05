package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCliConfigCmd(t *testing.T) {
	_, err := GetCliConfigCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigInitCmd(t *testing.T) {
	_, err := GetCliConfigInitCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigUpdateCmd(t *testing.T) {
	_, err := GetCliConfigUpdateCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigClearCmd(t *testing.T) {
	_, err := GetCliConfigClearCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigReadCmd(t *testing.T) {
	_, err := GetCliConfigReadCmd()
	assert.Nil(t, err)
}

func TestGetCliConfigEditCmd(t *testing.T) {
	_, err := GetCliConfigEditCmd()
	assert.Nil(t, err)
}

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
		assert.Equal(tC.expected, IsValidProfile(tC.profile), tC.profile)
	}
}
