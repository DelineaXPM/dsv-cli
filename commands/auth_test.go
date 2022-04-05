package cmd

import (
	"testing"

	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/vaultcli"

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
			httpClient.GetTokenStub = func() (response *auth.TokenResponse, apiError *errors.ApiError) {
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
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy Path",
			arg:         []string{},
			out:         nil,
			storeType:   "",
			expectedErr: nil,
		},
		{
			name:        "Error",
			arg:         []string{},
			out:         nil,
			storeType:   "",
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

			st := &fake.FakeStore{}

			st.WipeStub = func(s string) *errors.ApiError {
				return tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.StoreType, tt.storeType)

			_ = handleAuthClear(vcli, tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
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
		expectedErr *errors.ApiError
	}{
		{
			name:        "Happy Path",
			arg:         []string{},
			out:         []byte(`key1`),
			storeType:   "",
			list:        []string{"key1"},
			expectedErr: nil,
		},
		{
			name:        "Error",
			arg:         []string{},
			out:         nil,
			storeType:   "",
			list:        nil,
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

			st := &fake.FakeStore{}
			st.ListStub = func(s string) (strings []string, apiError *errors.ApiError) {
				return tt.list, tt.expectedErr
			}

			vcli, rerr := vaultcli.NewWithOpts(
				vaultcli.WithOutClient(outClient),
				vaultcli.WithStore(st),
			)
			if rerr != nil {
				t.Fatalf("Unexpected error during vaultCLI init: %v", err)
			}

			viper.Reset()
			viper.Set(cst.StoreType, tt.storeType)

			_ = handleAuthList(vcli, tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.out, data)
			} else {
				assert.Equal(t, tt.expectedErr, err)
			}

			viper.Set(cst.StoreType, "")
		})
	}
}
