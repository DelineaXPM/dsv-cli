package vaultcli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidProfile(t *testing.T) {
	testCases := []struct {
		profile string
		noErr   bool
	}{
		{profile: "profile1", noErr: true},
		{profile: "profile_name", noErr: true},
		{profile: "profile-name", noErr: true},
		{profile: "profile%name", noErr: true},
		{profile: "profilename'", noErr: true},
		{profile: "=", noErr: true},
		{profile: "profile name", noErr: false},
		{profile: " profilename", noErr: false},
		{profile: "profilename  ", noErr: false},
		{profile: `"profilename "`, noErr: false},
	}
	for _, tc := range testCases {
		got := IsValidProfile(tc.profile)
		if tc.noErr {
			assert.NoError(t, got, fmt.Sprintf("IsValidProfile(%s) should not return error", tc.profile))
		} else {
			assert.Error(t, got, fmt.Sprintf("IsValidProfile(%s) should return error", tc.profile))
		}
	}
}
