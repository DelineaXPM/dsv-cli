package cmd

import (
	"fmt"
	"testing"

	"github.com/DelineaXPM/dsv-cli/auth"
	cst "github.com/DelineaXPM/dsv-cli/constants"
	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/tests/fake"
	"github.com/DelineaXPM/dsv-cli/vaultcli"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetAuthCmd(t *testing.T) {
	_, err := GetAuthCmd()
	assert.Nil(t, err)
}

func TestGetAuthClearCmd(t *testing.T) {
	_, err := GetAuthClearCmd()
	assert.Nil(t, err)
}

func TestGetAuthListCmd(t *testing.T) {
	_, err := GetAuthListCmd()
	assert.Nil(t, err)
}

func TestGetAuthChangePasswordCmd(t *testing.T) {
	_, err := GetAuthChangePasswordCmd()
	assert.Nil(t, err)
}

func TestHandleAuth(t *testing.T) {
	testCase := []struct {
		name        string
		arg         []string
		in          *auth.TokenResponse
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy Path",
			arg:         []string{},
			in:          &auth.TokenResponse{Token: "token", RefreshToken: "refresh token"},
			out:         []byte(`{"accessToken":"token","tokenType":"","expiresIn":0,"refreshToken":"refresh token","granted":"0001-01-01T00:00:00Z"}`),
			expectedErr: nil,
		},
		{
			name:        "Error",
			arg:         []string{},
			in:          &auth.TokenResponse{Token: "token", RefreshToken: "refresh token"},
			out:         nil,
			expectedErr: errors.NewS("error"),
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err *errors.ApiError

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			httpClient := &fake.FakeAuthenticator{}
			httpClient.GetTokenStub = func() (*auth.TokenResponse, *errors.ApiError) {
				return tt.in, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(outClient),
				vaultcli.WithAuthenticator(httpClient),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()

			_ = handleAuth(vcli, tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func TestHandleAuthClear(t *testing.T) {
	testCase := []struct {
		name        string
		arg         []string
		out         []byte
		storeType   string
		expectedErr string
	}{
		{
			name:        "Happy Path",
			arg:         []string{},
			out:         nil,
			storeType:   "",
			expectedErr: "",
		},
		{
			name:        "Error",
			arg:         []string{},
			out:         nil,
			storeType:   "",
			expectedErr: "one two",
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
			}

			st := &fake.FakeStore{}
			st.WipeStub = func(s string) error { return fmt.Errorf(tt.expectedErr) }

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Reset()
			viper.Set(cst.StoreType, tt.storeType)

			err := handleAuthClear(vcli, tt.arg)
			if tt.expectedErr == "" {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}

func TestHandleAuthList(t *testing.T) {
	testCase := []struct {
		name        string
		arg         []string
		out         []byte
		storeType   string
		list        []string
		expectedErr string
	}{
		{
			name:        "Happy Path",
			arg:         []string{},
			out:         []byte(`key1`),
			storeType:   "",
			list:        []string{"key1"},
			expectedErr: "",
		},
		{
			name:        "Error",
			arg:         []string{},
			out:         nil,
			storeType:   "",
			list:        nil,
			expectedErr: "one two",
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte

			outClient := &fake.FakeOutClient{}
			outClient.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
			}

			st := &fake.FakeStore{}
			st.ListStub = func(s string) ([]string, error) {
				if tt.expectedErr != "" {
					return nil, fmt.Errorf(tt.expectedErr)
				}
				return tt.list, nil
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", rerr)
			}

			viper.Reset()
			viper.Set(cst.StoreType, tt.storeType)

			err := handleAuthList(vcli, tt.arg)
			if tt.expectedErr == "" {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Contains(t, err.Error(), tt.expectedErr)
			}

			viper.Set(cst.StoreType, "")
		})
	}
}
