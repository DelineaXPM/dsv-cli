//go:build !codeanalysis
// +build !codeanalysis

package cmd

import (
	"testing"

	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/vaultcli"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBYOKCmd(t *testing.T) {
	_, err := GetBYOKCmd()
	assert.Nil(t, err)
}

func TestBYOKUpdateCmd(t *testing.T) {
	_, err := GetBYOKUpdateCmd()
	assert.Nil(t, err)
}

func TestByokUpdateCmd(t *testing.T) {
	testCase := []struct {
		name         string
		primaryKey   string
		secondaryKey string
		apiOut       []byte
		apiErr       *errors.ApiError
		wantOut      []byte
		wantErr      *errors.ApiError
	}{
		{
			name:         "success",
			primaryKey:   "key",
			secondaryKey: "key",
			apiOut:       []byte(`{"response":"success"}`),
			wantOut:      []byte(`{"response":"success"}`),
		},
		{
			name:         "API error",
			primaryKey:   "key",
			secondaryKey: "key",
			apiErr:       errors.NewS(`{"error":"message"}`),
			wantErr:      errors.NewS(`{"error":"message"}`),
		},
		{
			name:         "missing --primary-key",
			secondaryKey: "key",
			wantErr:      errors.NewS("error: must specify primary-key"),
		},
		{
			name:       "missing --secondary-key",
			primaryKey: "key",
			apiOut:     []byte(`{"response":"success"}`),
			wantOut:    []byte(`{"response":"success"}`),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeClient{}
			httpClient.DoRequestStub = func(s string, s2 string, i interface{}) (bytes []byte, apiError *errors.ApiError) {
				return tc.apiOut, tc.apiErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithHTTPClient(httpClient),
				vaultcli.WithOutClient(outClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Reset()
			viper.Set(cst.PrimaryKey, tc.primaryKey)
			viper.Set(cst.SecondaryKey, tc.secondaryKey)

			_ = handleBYOKUpdateCmd(vcli, []string{})

			if tc.wantErr == nil {
				assert.Equal(t, tc.wantOut, data)
			} else {
				assert.Equal(t, tc.wantErr, err)
			}
		})
	}
}
