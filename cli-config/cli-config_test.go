package cliconfig

import (
	"testing"
	"thy/errors"

	"github.com/stretchr/testify/assert"
)

func TestGetSecureSetting(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		profile       string
		expectedError error
		expectedVal   string
	}{
		{"missing-key", "", "some-profile", errors.NewS("key cannot be empty"), ""},
		{"empty-value", "hello", "some-profile", nil, ""},
		{"missing-profile", "hello", "", errors.NewS("profile cannot be empty"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := GetSecureSettingForProfile(tt.key, tt.profile)

			if tt.expectedError != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			}
			if err == nil {
				assert.Equal(t, tt.expectedVal, val)
			}
		})

	}
}

func TestGetFlagBeforeParse(t *testing.T) {
	testCases := []struct {
		name           string
		flag           string
		args           []string
		expectedResult string
	}{

		{
			name:           "Correct result for short form of the flag (no conflict with '-c' suffix of the argument value)",
			flag:           "config",
			args:           []string{"secret-c", "--profile", "testprofile", "-c", "some_config_value"},
			expectedResult: "some_config_value",
		},

		{
			name:           "Correct result for the config flag in the beginning of the args",
			flag:           "config",
			args:           []string{"-c", "some_config_value", "another_arg"},
			expectedResult: "some_config_value",
		},

		{
			name: "Correct result for short form of the flag (no conflict with '-c' suffix of the argument " +
				"value). With '=' symbol.",
			flag:           "config",
			args:           []string{"secret-c", "--profile", "testprofile", "-c=some_config_value"},
			expectedResult: "some_config_value",
		},

		{
			name:           "Correct result for the config flag in the beginning of the args. With '=' symbol.",
			flag:           "config",
			args:           []string{"-c=some_config_value", "another_arg"},
			expectedResult: "some_config_value",
		},

		// long form of the flag
		{
			name:           "Long form of the flag. Correct result for short form of the flag (no conflict with '--config' suffix of the argument value)",
			flag:           "config",
			args:           []string{"secret--config", "--profile", "testprofile", "--config", "some_config_value"},
			expectedResult: "some_config_value",
		},

		{
			name:           "Long form of the flag. Correct result for the config flag in the beginning of the args",
			flag:           "config",
			args:           []string{"--config", "some_config_value", "another_arg"},
			expectedResult: "some_config_value",
		},

		{
			name: "Long form of the flag. Correct result for short form of the flag (no conflict with '--config' " +
				"suffix of the argument value). With '=' symbol.",
			flag:           "config",
			args:           []string{"secret--config", "--profile", "testprofile", "--config=some_config_value"},
			expectedResult: "some_config_value",
		},

		{
			name: "Long form of the flag. Correct result for the config flag in the beginning of the args. " +
				"With '=' symbol.",
			flag:           "config",
			args:           []string{"--config=some_config_value", "another_arg"},
			expectedResult: "some_config_value",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flagValue := GetFlagBeforeParse(testCase.flag, testCase.args)
			assert.Equal(t, testCase.expectedResult, flagValue)
		})
	}

}
