package cmd

import (
	"testing"

	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetWhoAmICmd(t *testing.T) {
	_, err := GetWhoAmICmd()
	assert.Nil(t, err)
}

func TestGetEvaluateFlagCmd(t *testing.T) {
	_, err := GetEvaluateFlagCmd()
	assert.Nil(t, err)
}

func TestHandleEvaluateFlag(t *testing.T) {
	testCase := []struct {
		name string
		args string
	}{
		{"Happy path 1", "arg.one"},
		{"Happy path 2", "arg-one"},
		{"Happy path 3", "arg_one"},
		{"Happy path 4", "--arg.one"},
		{"Happy path 5", "--arg-one"},
		{"Happy path 6", "--arg_one"},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Set("arg.one", "arg value")

			_ = handleEvaluateFlag(vcli, []string{tt.args})
			assert.Equal(t, []byte("arg value"), data)

			viper.Set("arg.one", "")
		})
	}
}
