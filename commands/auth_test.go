package cmd

import (
	e "errors"
	"testing"
	"thy/auth"
	cst "thy/constants"
	"thy/errors"
	"thy/fake"
	"thy/store"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHandleAuth(t *testing.T) {
	testCase := []struct {
		name        string
		arg         []string
		in          *auth.TokenResponse
		out         []byte
		expectedErr *errors.ApiError
	}{
		{
			"Happy Path",
			[]string{},
			&auth.TokenResponse{Token: "token", RefreshToken: "refresh token"},
			[]byte(`{"accessToken":"token","tokenType":"","expiresIn":0,"refreshToken":"refresh token","granted":"0001-01-01T00:00:00Z"}`),
			nil,
		},
		{
			"Error",
			[]string{},
			&auth.TokenResponse{Token: "token", RefreshToken: "refresh token"},
			nil,
			errors.New(e.New("error")),
		},
	}

	cmd, err := GetAuthCmd()
	help := cmd.Help()
	assert.Contains(t, help, "Authenticate with")
	assert.Nil(t, err)
	viper.Set(cst.Version, "v1")

	for _, tt := range testCase {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			tok := &fake.FakeAuthenticator{}
			tok.GetTokenStub = func() (response *auth.TokenResponse, apiError *errors.ApiError) {
				return tt.in, tt.expectedErr
			}
			newAuthenticatorFunc := func() auth.Authenticator {
				return tok
			}
			authCmd := &AuthCommand{acmd, newAuthenticatorFunc, store.GetStore, nil}
			_ = authCmd.handleAuth(tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
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
			"Happy Path",
			[]string{},
			nil,
			"",
			nil,
		},
		{
			"Error",
			[]string{},
			nil,
			"",
			errors.New(e.New("error")),
		},
	}

	_, err := GetAuthClearCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			st := &fake.FakeStore{}

			st.WipeStub = func(s string) *errors.ApiError {
				return tt.expectedErr
			}

			viper.Set(cst.StoreType, tt.storeType)
			authCmd := &AuthCommand{acmd, nil, store.GetStore, nil}
			authCmd.getStore = func(stString string) (i store.Store, apiError *errors.ApiError) {
				return st, nil
			}
			_ = authCmd.handleAuthClear(tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}
			viper.Set(cst.StoreType, "")
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
			"Happy Path",
			[]string{},
			[]byte(`key1`),
			"",
			[]string{"key1"},
			nil,
		},
		{
			"Error",
			[]string{},
			nil,
			"",
			nil,
			errors.New(e.New("error")),
		},
	}

	_, err := GetAuthListCmd()
	assert.Nil(t, err)

	for _, tt := range testCase {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			acmd := &fake.FakeOutClient{}
			var data []byte
			var err *errors.ApiError
			acmd.WriteResponseStub = func(bytes []byte, apiError *errors.ApiError) {
				data = bytes
				err = apiError
			}

			st := &fake.FakeStore{}

			st.ListStub = func(s string) (strings []string, apiError *errors.ApiError) {
				return tt.list, tt.expectedErr
			}

			viper.Set(cst.StoreType, tt.storeType)
			authCmd := &AuthCommand{acmd, nil, store.GetStore, nil}
			authCmd.getStore = func(stString string) (i store.Store, apiError *errors.ApiError) {
				return st, nil
			}
			_ = authCmd.handleAuthList(tt.arg)
			if tt.expectedErr == nil {
				assert.Equal(t, data, tt.out)
			} else {
				assert.Equal(t, err, tt.expectedErr)
			}

			viper.Set(cst.StoreType, "")
		})
	}
}
